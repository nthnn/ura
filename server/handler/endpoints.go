package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/nthnn/ura/util"
)

var (
	timeoutMinute time.Duration = 60

	errMethodNotAllowed                   = "Method Not Allowed"
	errInvalidRequest                     = "Invalid request body"
	errInvalidUsername                    = "Username cannot contain punctuations except underscore"
	errInvalidSignupCredentials           = "Invalid credentials"
	errInternalErrorOccurred              = "Internal error occurred"
	errUserStillActive                    = "Account has recent activity"
	errPaymentRequestNotFound             = "Payment request not found"
	errCannotPayOwnAccount                = "Cannot process payment to self"
	errInsufficientFunds                  = "Insufficient funds"
	errPaymentExceeds100kUro              = "Payment amount exceeds 100k uro"
	errExceededReceivedFunds              = "Recipient received funds limit exceeded in past 2 business days"
	errPaymentAlreadyProcessed            = "Payment request already processed"
	errInvalidAmountValue                 = "Invalid amount value"
	errInvalidWithdrawAmount              = "Withdraw amount cannot be zero or negative value"
	errInvalidWithdrawAmountExceeds50kUro = "Withdraw amount must be less than 50k uro"
	errInvalidCashInAmount                = "Cash in amount cannot be zero or negative value"
	errCashInAmountExceeds100kUro         = "Cash in amount must be less than 100k uro"
	errCashInEveryIn12Hours               = "Cash in allowed only every 12 hours"
	errInvalidLoginCredentials            = "Invalid log-in credentials"
)

func UserCreate(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateUsername(req.Username) {
			util.WriteJSONError(w, errInvalidUsername)
			return
		}

		if !util.ValidateEmail(req.Email) {
			util.WriteJSONError(w, errInvalidSignupCredentials)
			return
		}

		if !util.IsValidSHA512(req.Password) {
			util.WriteJSONError(w, errInvalidSignupCredentials)
			return
		}

		var count int
		err = db.QueryRow(
			"SELECT COUNT(*) FROM users WHERE username = ? OR email = ?",
			req.Username, req.Email,
		).Scan(&count)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		if count > 0 {
			util.WriteJSONError(w, errInvalidSignupCredentials)
			return
		}

		identifier, err := util.GenerateRandomIdentifier(128)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		securityCode, err := util.GenerateRandomIdentifier(128)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		stmt, err := db.Prepare(
			"INSERT INTO users (username, email, password, identifier, security_code, balance_ura, created_at) " +
				"VALUES (?, ?, ?, ?, ?, 0, ?)",
		)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}
		defer stmt.Close()

		now := time.Now().UTC().Format(time.RFC3339)
		_, err = stmt.Exec(
			req.Username,
			req.Email,
			req.Password,
			identifier,
			securityCode,
			now,
		)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]interface{}{
			"status": "ok",
		})
	}
}

func UserDelete(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		user, authErr := authenticate(db, r)
		if authErr != "" {
			util.WriteJSONError(w, authErr)
			return
		}

		var lastActivity string
		err := db.QueryRow(
			"SELECT MAX(created_at) FROM transactions WHERE user_id = ?",
			user.ID,
		).Scan(&lastActivity)

		if err != nil {
			lastActivity = ""
		}

		if lastActivity != "" {
			lastTime, err := time.Parse(time.RFC3339, lastActivity)

			if err == nil && time.Since(lastTime) < (30*24*time.Hour) {
				util.WriteJSONError(w, errUserStillActive)
				return
			}
		}

		stmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(user.ID)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]string{"status": "ok"})
	}
}

func PaymentProcess(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		payer, authErr := authenticate(db, r)
		if authErr != "" {
			util.WriteJSONError(w, authErr)
			return
		}

		var req struct {
			TransactionID string `json:"transaction_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if req.TransactionID == "" {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateTransactionID(req.TransactionID) {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		var amount float64
		var recipientID int64

		err := db.QueryRow(
			`SELECT amount, user_id FROM transactions
			 WHERE transaction_id = ? AND category = 'payment_request'`,
			req.TransactionID,
		).Scan(&amount, &recipientID)

		if err != nil {
			util.WriteJSONError(w, errPaymentRequestNotFound)
			return
		}

		if recipientID == payer.ID {
			util.WriteJSONError(w, errCannotPayOwnAccount)
			return
		}

		if payer.BalanceUra < amount {
			util.WriteJSONError(w, errInsufficientFunds)
			return
		}

		if amount > 100000 {
			util.WriteJSONError(w, errPaymentExceeds100kUro)
			return
		}
		twoBusinessDaysAgo := time.Now().Add(-48 * time.Hour)

		var receivedSum float64
		err = db.QueryRow(
			`SELECT COALESCE(SUM(amount), 0) FROM transactions 
			 WHERE user_id = ? AND created_at > ? AND category = 'incoming' AND processed = 1`,
			recipientID, twoBusinessDaysAgo.Format(time.RFC3339),
		).Scan(&receivedSum)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		if receivedSum >= 50000 {
			util.WriteJSONError(w, errExceededReceivedFunds)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		res, err := tx.Exec(
			`UPDATE transactions 
			 SET processed = 1
			 WHERE transaction_id = ? AND category = 'payment_request' AND processed = 0`,
			req.TransactionID,
		)

		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, errInternalErrorOccurred)

			return
		}

		if rowsAffected == 0 {
			tx.Rollback()
			util.WriteJSONError(w, errPaymentAlreadyProcessed)

			return
		}

		if _, err = tx.Exec(
			"UPDATE users SET balance_ura = balance_ura - ? WHERE id = ?",
			amount, payer.ID,
		); err != nil {
			tx.Rollback()
			util.WriteJSONError(w, errInternalErrorOccurred)

			return
		}

		if _, err = tx.Exec(
			"UPDATE users SET balance_ura = balance_ura + ? WHERE id = ?",
			amount, recipientID,
		); err != nil {
			tx.Rollback()
			util.WriteJSONError(w, errInternalErrorOccurred)

			return
		}

		if _, err = tx.Exec(
			`INSERT INTO transactions (transaction_id, user_id, category, amount, created_at, processed)
			 VALUES (?, ?, 'incoming', ?, ?, 1)`,
			req.TransactionID, recipientID, amount, now,
		); err != nil {
			tx.Rollback()
			util.WriteJSONError(w, errInternalErrorOccurred)

			return
		}

		if err = tx.Commit(); err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":         "ok",
			"transaction_id": req.TransactionID,
		})
	}
}

func PaymentRequest(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		user, authErr := authenticate(db, r)
		if authErr != "" {
			util.WriteJSONError(w, authErr)
			return
		}

		var req struct {
			Amount string `json:"amount"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateNumbers(req.Amount) {
			util.WriteJSONError(w, errInvalidAmountValue)
			return
		}

		amount, err := strconv.ParseFloat(req.Amount, 64)
		if err != nil {
			util.WriteJSONError(w, errInvalidAmountValue)
			return
		}

		if amount > 100000 {
			util.WriteJSONError(w, errPaymentExceeds100kUro)
			return
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt, err := db.Prepare(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at, processed) " +
				"VALUES (?, ?, 'payment_request', ?, ?, 0)",
		)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(transactionID, user.ID, amount, now)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":         "ok",
			"transaction_id": transactionID,
		})
	}
}

func Withdraw(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		user, authErr := authenticate(db, r)
		if authErr != "" {
			util.WriteJSONError(w, authErr)
			return
		}

		var req struct {
			Amount string `json:"amount"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateNumbers(req.Amount) {
			util.WriteJSONError(w, errInvalidAmountValue)
			return
		}

		amount, err := strconv.ParseFloat(req.Amount, 64)
		if err != nil {
			util.WriteJSONError(w, errInvalidAmountValue)
			return
		}

		if amount <= 0 {
			util.WriteJSONError(w, errInvalidWithdrawAmount)
			return
		} else if amount >= 50000 {
			util.WriteJSONError(w, errInvalidWithdrawAmountExceeds50kUro)
			return
		}
		twoBusinessDaysAgo := time.Now().Add(-48 * time.Hour)

		var receivedSum float64
		err = db.QueryRow(
			"SELECT COALESCE(SUM(amount),0) FROM transactions "+
				"WHERE user_id = ? AND created_at > ? AND category = 'incoming'",
			user.ID,
			twoBusinessDaysAgo.Format(time.RFC3339),
		).Scan(&receivedSum)

		if err == nil && receivedSum >= 50000 {
			util.WriteJSONError(
				w,
				"Cannot withdraw after receiving 50k uro in past 2 business days",
			)
			return
		}

		var balanceUra float64
		err = db.QueryRow(
			"SELECT balance_ura FROM users WHERE id = ?",
			user.ID,
		).Scan(&balanceUra)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		if balanceUra < amount {
			util.WriteJSONError(w, errInsufficientFunds)
			return
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt2, err := db.Prepare(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at, processed) " +
				"VALUES (?, ?, 'withdraw', ?, ?, 0)",
		)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		_, err = stmt2.Exec(transactionID, user.ID, amount, now)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		stmt2.Close()
		util.WriteJSON(w, map[string]string{
			"status":         "ok",
			"transaction_id": transactionID,
		})
	}
}

func CashIn(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		user, authErr := authenticate(db, r)
		if authErr != "" {
			util.WriteJSONError(w, authErr)
			return
		}

		var req struct {
			Amount string `json:"amount"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateNumbers(req.Amount) {
			util.WriteJSONError(w, errInvalidAmountValue)
			return
		}

		amount, err := strconv.ParseFloat(req.Amount, 64)
		if err != nil {
			util.WriteJSONError(w, errInvalidAmountValue)
			return
		}

		if amount <= 0 {
			util.WriteJSONError(w, errInvalidCashInAmount)
			return
		} else if amount >= 100000 {
			util.WriteJSONError(w, errCashInAmountExceeds100kUro)
			return
		}

		var lastCashInStr string
		err = db.QueryRow(
			"SELECT MAX(created_at) FROM transactions "+
				"WHERE user_id = ? AND category = 'cashin'",
			user.ID,
		).Scan(&lastCashInStr)

		if err == nil && lastCashInStr != "" {
			lastCashIn, err := time.Parse(time.RFC3339, lastCashInStr)

			if err == nil && time.Since(lastCashIn) < 12*time.Hour {
				util.WriteJSONError(w, errCashInEveryIn12Hours)
				return
			}
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt2, err := db.Prepare(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at, processed) " +
				"VALUES (?, ?, 'cashin', ?, ?, false)",
		)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		_, err = stmt2.Exec(transactionID, user.ID, amount, now)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		stmt2.Close()
		util.WriteJSON(w, map[string]string{
			"status":         "ok",
			"transaction_id": transactionID,
		})
	}
}

func UserLogin(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateUsername(req.Username) {
			util.WriteJSONError(w, errInvalidLoginCredentials)
			return
		}

		if !util.IsValidSHA512(req.Password) {
			util.WriteJSONError(w, errInvalidLoginCredentials)
			return
		}

		var user User
		var createdAtStr string

		err = db.QueryRow(
			"SELECT id, username, email, identifier, security_code, balance_ura, created_at "+
				"FROM users WHERE username = ? AND password = ?",
			req.Username,
			req.Password,
		).Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Identifier,
			&user.SecurityCode,
			&user.BalanceUra,
			&createdAtStr,
		)

		if err != nil {
			util.WriteJSONError(w, errInvalidLoginCredentials)
			return
		}

		sessionToken, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		expiresAt := time.Now().Add(timeoutMinute * time.Minute).UTC().Format(time.RFC3339)
		stmt, err := db.Prepare("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)")

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(sessionToken, user.ID, expiresAt)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":        "ok",
			"session_token": sessionToken,
			"security_code": user.SecurityCode,
		})
	}
}

func UserFetchInfo(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		user, authErr := authenticate(db, r)
		if authErr != "" {
			util.WriteJSONError(w, authErr)
			return
		}

		rows, err := db.Query(
			"SELECT transaction_id, category, amount, created_at, processed "+
				"FROM transactions WHERE user_id = ?",
			user.ID,
		)

		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}
		defer rows.Close()

		var transactions []map[string]interface{}
		for rows.Next() {
			var tid, category, createdAt string
			var processed int
			var amount float64

			err = rows.Scan(&tid, &category, &amount, &createdAt, &processed)
			if err != nil {
				util.WriteJSONError(w, errInternalErrorOccurred)
				return
			}

			transactions = append(transactions, map[string]interface{}{
				"transaction_id": tid,
				"category":       category,
				"amount":         amount,
				"created_at":     createdAt,
				"processed":      processed,
			})
		}

		if err = rows.Err(); err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]interface{}{
			"user":         user,
			"transactions": transactions,
		})
	}
}

func UserLogout(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		sessionToken := r.Header.Get("X-Session-Token")
		if sessionToken == "" {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		stmt, err := db.Prepare("DELETE FROM sessions WHERE token = ?")
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		_, err = stmt.Exec(sessionToken)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status": "ok",
		})
	}
}

func ValidateSession(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		sessionToken := r.Header.Get("X-Session-Token")
		if sessionToken == "" {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		if !util.ValidateSessionToken(sessionToken) {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		var userID int64
		var expiresAtStr string

		err := db.QueryRow(
			"SELECT user_id, expires_at FROM sessions WHERE token = ?",
			sessionToken,
		).Scan(&userID, &expiresAtStr)

		if err != nil {
			util.WriteJSONError(w, errInvalidRequest)
			return
		}

		expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			util.WriteJSONError(w, errInternalErrorOccurred)
			return
		}

		hasExpired := time.Now().After(expiresAt)
		if hasExpired {
			if _, err := db.Exec(
				"DELETE FROM sessions WHERE token = ?",
				sessionToken,
			); err != nil {
				util.WriteJSONError(w, errInternalErrorOccurred)
				return
			}
		}

		util.WriteJSON(w, map[string]interface{}{
			"status":     "ok",
			"user_id":    strconv.FormatInt(userID, 10),
			"expired":    strconv.FormatBool(hasExpired),
			"expires_at": expiresAtStr,
		})
	}
}
