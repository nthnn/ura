//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"syscall/js"
	"time"
)

type Transaction struct {
	Amount        float64 `json:"amount"`
	Category      string  `json:"category"`
	CreatedAt     string  `json:"created_at"`
	TransactionID string  `json:"transaction_id"`
	Processed     int     `json:"processed"`
}

type User struct {
	ID         int     `json:"id"`
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	Identifier string  `json:"identifier"`
	BalanceUra float64 `json:"balance_ura"`
	CreatedAt  string  `json:"created_at"`
}

type Response struct {
	Transactions []Transaction `json:"transactions"`
	User         User          `json:"user"`
}

var previousHash string

func renderTransaction(id string, transaction Transaction) {
	htmlContents := `
	<tr>
		<td class="p-2">%s</td>
		<td class="p-2">%s</td>
		<td class="p-2">%g</td>
		<td class="p-2 desktop-only" alt="%s">%s</td>
		<td class="p-2 desktop-only">%s</td>
	</tr>
	`

	status := "✓"
	if transaction.Processed != 1 {
		status = "×"
	}

	formattedTimestamp := transaction.CreatedAt
	if parsedTime, err := time.Parse(time.RFC3339, transaction.CreatedAt); err == nil {
		formattedTimestamp = parsedTime.Format("02/01/2006 15:04:05 MST")
	}

	js.Global().Get("document").Call(
		"getElementById",
		id,
	).Call(
		"insertAdjacentHTML",
		"beforeend",
		fmt.Sprintf(
			htmlContents,
			capitalizeFirst(transaction.Category),
			formattedTimestamp,
			transaction.Amount,
			transaction.TransactionID,
			transaction.TransactionID[:12],
			status,
		),
	)
}

func parseJSON(data []byte) Response {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	return resp
}

func fetchInformation() (Response, string, error) {
	status, _, content := sendPost(
		"/api/user/info",
		map[string]string{},
		map[string]interface{}{
			"X-Session-Token": getSessionKey("session_token"),
			"X-Security-Code": getSessionKey("security_code"),
		},
	)

	if status != 200 {
		return Response{}, "", errors.New("server error")
	}

	var data Response
	err := json.Unmarshal([]byte(content), &data)

	if err != nil {
		return Response{}, "", errors.New("failed parsing response data")
	}

	return data, toSHA512(content), nil
}

func loadInitialInformation() {
	data, hash, err := fetchInformation()
	if err == nil {
		if hash == previousHash {
			return
		}

		username := data.User.Username
		previousHash = hash

		document.Call(
			"getElementById",
			"card-holder",
		).Set("innerHTML", username)

		document.Call(
			"getElementById",
			"card-identification",
		).Set("innerHTML", data.User.Identifier)

		document.Call(
			"getElementById",
			"overview-username",
		).Set("innerHTML", username)

		document.Call(
			"getElementById",
			"credit-amount",
		).Set("innerHTML", numberWithCommas(data.User.BalanceUra))

		sort.Slice(data.Transactions, func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339, data.Transactions[i].CreatedAt)
			t2, _ := time.Parse(time.RFC3339, data.Transactions[j].CreatedAt)

			return t1.After(t2)
		})

		transactionTable := document.Call(
			"getElementById",
			"transaction-table",
		).Get("classList")
		noTransactionTable := document.Call(
			"getElementById",
			"no-transactions",
		).Get("classList")

		if len(data.Transactions) != 0 {
			transactionTable.Call("remove", "d-none")
			transactionTable.Call("add", "d-block")

			noTransactionTable.Call("remove", "d-block")
			noTransactionTable.Call("add", "d-none")

			document.Call(
				"getElementById",
				"transactions",
			).Set("innerHTML", "")

			for _, transaction := range data.Transactions {
				renderTransaction("transactions", transaction)
			}

			document.Call(
				"getElementById",
				"download-transaction-btn",
			).Call(
				"addEventListener",
				"click",
				js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					generateTransactionPDF(data.Transactions)
					return nil
				}),
			)
		} else {
			transactionTable.Call("remove", "d-block")
			transactionTable.Call("add", "d-none")

			noTransactionTable.Call("remove", "d-none")
			noTransactionTable.Call("add", "d-block")
		}
	}
}

func updateInformation() {
	go func() {
		ticker := time.NewTicker(6 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			loadInitialInformation()
		}
	}()
}
