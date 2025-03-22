//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
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
var downloadCallback js.Func

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

	tbody := js.Global().Get("document").Call(
		"getElementById",
		id,
	)
	if !tbody.IsNull() && !tbody.IsUndefined() {
		tidDisplay := transaction.TransactionID
		if len(tidDisplay) > 12 {
			tidDisplay = tidDisplay[:12]
		}

		tbody.Call(
			"insertAdjacentHTML",
			"beforeend",
			fmt.Sprintf(
				htmlContents,
				html.EscapeString(capitalizeFirst(transaction.Category)),
				formattedTimestamp,
				transaction.Amount,
				html.EscapeString(transaction.TransactionID),
				html.EscapeString(tidDisplay),
				status,
			),
		)
	}
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

		overviewUsername := document.Call(
			"getElementById",
			"overview-username",
		)
		cardHolder := document.Call(
			"getElementById",
			"card-holder",
		)
		identifier := document.Call(
			"getElementById",
			"card-identification",
		)
		creditAmount := document.Call(
			"getElementById",
			"credit-amount",
		)

		if !overviewUsername.IsNull() && !overviewUsername.IsUndefined() {
			overviewUsername.Set(
				"innerHTML",
				html.EscapeString(username),
			)
		}

		if !cardHolder.IsNull() && !cardHolder.IsUndefined() {
			cardHolder.Set(
				"innerHTML",
				html.EscapeString(username),
			)
		}

		if !identifier.IsNull() && !identifier.IsUndefined() {
			identifier.Set(
				"innerHTML",
				html.EscapeString(data.User.Identifier),
			)
		}

		if !creditAmount.IsNull() && !creditAmount.IsUndefined() {
			creditAmount.Set(
				"innerHTML",
				html.EscapeString(numberWithCommas(data.User.BalanceUra)),
			)
		}

		sort.Slice(data.Transactions, func(i, j int) bool {
			t1, err := time.Parse(time.RFC3339, data.Transactions[i].CreatedAt)
			if err != nil {
				return false
			}

			t2, err := time.Parse(time.RFC3339, data.Transactions[j].CreatedAt)
			if err != nil {
				return false
			}

			return t1.After(t2)
		})

		transactionTable := document.Call(
			"getElementById",
			"transaction-table",
		).Get("classList")

		if transactionTable.IsNull() || transactionTable.IsUndefined() {
			return
		}

		noTransactionTable := document.Call(
			"getElementById",
			"no-transactions",
		).Get("classList")

		if noTransactionTable.IsNull() || noTransactionTable.IsUndefined() {
			return
		}

		if len(data.Transactions) != 0 {
			transactionTable.Call("remove", "d-none")
			transactionTable.Call("add", "d-block")

			noTransactionTable.Call("remove", "d-block")
			noTransactionTable.Call("add", "d-none")

			transactions := document.Call(
				"getElementById",
				"transactions",
			)

			if !transactions.IsNull() && !transactions.IsUndefined() {
				transactions.Set("innerHTML", "")
			}

			for _, transaction := range data.Transactions {
				renderTransaction("transactions", transaction)
			}

			downloadBtn := document.Call(
				"getElementById",
				"download-transaction-btn",
			)
			if !downloadBtn.IsNull() && !downloadBtn.IsUndefined() {
				if !downloadCallback.IsNull() && !downloadCallback.IsUndefined() {
					downloadBtn.Call("removeEventListener", "click", downloadCallback)
					downloadCallback.Release()
				}

				downloadCallback = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					generateTransactionPDF(data.Transactions)
					return nil
				})

				downloadBtn.Call(
					"addEventListener",
					"click",
					downloadCallback,
				)
			}
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
