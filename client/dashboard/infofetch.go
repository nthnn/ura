//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"
)

type Transaction struct {
	Amount        float64 `json:"amount"`
	Category      string  `json:"category"`
	CreatedAt     string  `json:"created_at"`
	TransactionID string  `json:"transaction_id"`
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
	Loans        interface{}   `json:"loans"`
	Transactions []Transaction `json:"transactions"`
	User         User          `json:"user"`
}

func parseJSON(data []byte) Response {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	return resp
}

func fetchInformation() (Response, error) {
	status, _, content := sendPost(
		"/api/user/info",
		map[string]string{},
		map[string]interface{}{
			"X-Session-Token": getSessionKey("session_token"),
			"X-Security-Code": getSessionKey("security_code"),
		},
	)

	if status != 200 {
		return Response{}, errors.New("server error")
	}

	var data Response
	err := json.Unmarshal([]byte(content), &data)

	if err != nil {
		return Response{}, errors.New("failed parsing response data")
	}

	return data, nil
}

func loadInitialInformation() {
	data, err := fetchInformation()
	if err == nil {
		username := data.User.Username

		document.Call(
			"getElementById",
			"card-holder",
		).Set("value", username)

		document.Call(
			"getElementById",
			"card-identification",
		).Set("value", data.User.Identifier)

		document.Call(
			"getElementById",
			"overview-username",
		).Set("innerHTML", username)

		document.Call(
			"getElementById",
			"credit-amount",
		).Set("innerHTML", numberWithCommas(data.User.BalanceUra))
	}
}

func updateInformation() {
	go func() {
		ticker := time.NewTicker(12 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			loadInitialInformation()
		}
	}()
}
