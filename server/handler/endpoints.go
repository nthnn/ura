package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nthnn/ura/util"
)

var (
	timeoutMinute time.Duration = 60
)

func UserCreate(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if !util.IsValidSHA512(req.Password) {
			util.WriteJSONError(w, "Password is not a valid SHA-512", http.StatusBadRequest)
			return
		}

		identifier, err := util.GenerateRandomIdentifier(128)
		if err != nil {
			util.WriteJSONError(w, "Error generating identifier", http.StatusInternalServerError)
			return
		}

		securityCode, err := util.GenerateRandomIdentifier(128)
		if err != nil {
			util.WriteJSONError(w, "Error generating security code", http.StatusInternalServerError)
			return
		}

		stmt, err := db.Prepare(
			"INSERT INTO users (username, email, password, identifier, security_code, balance_ura, created_at) " +
				"VALUES (?, ?, ?, ?, ?, 0, ?)",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
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
			util.WriteJSONError(w, "Database insert error", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]interface{}{
			"status": "200",
		})
	}
}

func UserDelete(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var lastActivity string
		err = db.QueryRow(
			"SELECT MAX(created_at) FROM ("+
				"SELECT created_at FROM transactions "+
				"WHERE user_id = ? "+
				"UNION SELECT created_at FROM loans "+
				"WHERE debtor_id = ? OR creditor_id = ?"+
				")",
			user.ID,
			user.ID,
			user.ID,
		).Scan(&lastActivity)

		if err != nil {
			lastActivity = ""
		}

		if lastActivity != "" {
			lastTime, err := time.Parse(time.RFC3339, lastActivity)

			if err == nil && time.Since(lastTime) < (24*30*24*time.Hour) {
				util.WriteJSONError(w, "User has recent activity", http.StatusBadRequest)
				return
			}
		}

		var totalUroLoans int64
		var totalUraLoans int64

		err = db.QueryRow(
			"SELECT COALESCE(SUM(amount),0) FROM loans "+
				"WHERE debtor_id = ? AND loan_type = 'uro' AND status = 'accepted'",
			user.ID,
		).Scan(&totalUroLoans)

		if err != nil {
			util.WriteJSONError(w, "Error checking loans", http.StatusInternalServerError)
			return
		}

		err = db.QueryRow(
			"SELECT COALESCE(SUM(amount),0) FROM loans "+
				"WHERE debtor_id = ? AND loan_type = 'ura' AND status = 'accepted'",
			user.ID,
		).Scan(&totalUraLoans)

		if err != nil {
			util.WriteJSONError(w, "Error checking loans", http.StatusInternalServerError)
			return
		}

		if totalUroLoans > 5 || totalUraLoans > 0 {
			util.WriteJSONError(w, "User does not meet deletion criteria", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(user.ID)
		if err != nil {
			util.WriteJSONError(w, "Error deleting user", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{"status": "user deleted"})
	}
}

func LoanRequest(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		debtor, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			CreditorIdentifier string  `json:"creditor_identifier"`
			Amount             float64 `json:"amount"`
			LoanType           string  `json:"loan_type"`
			TimeSpanDays       int     `json:"timespan_days"`
			PaymentType        string  `json:"payment_type"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var creditorID int64
		err = db.QueryRow(
			"SELECT id FROM users WHERE identifier = ?",
			req.CreditorIdentifier,
		).Scan(&creditorID)

		if err != nil {
			util.WriteJSONError(w, "Creditor not found", http.StatusBadRequest)
			return
		}

		loanID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, "Error generating loan ID", http.StatusInternalServerError)
			return
		}

		stmt, err := db.Prepare(
			"INSERT INTO loans " +
				"(loan_id, debtor_id, creditor_id, amount, loan_type, " +
				"timespan, payment_type, status, created_at) " +
				"VALUES (?, ?, ?, ?, ?, ?, ?, 'pending', ?)",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		now := time.Now().UTC().Format(time.RFC3339)
		_, err = stmt.Exec(
			loanID,
			debtor.ID,
			creditorID,
			req.Amount,
			req.LoanType,
			req.TimeSpanDays,
			req.PaymentType,
			now,
		)

		if err != nil {
			util.WriteJSONError(w, "Error creating loan request", http.StatusInternalServerError)
			return
		}

		message := fmt.Sprintf("Loan request %s from user %s", loanID, debtor.Identifier)
		stmt2, err := db.Prepare(
			"INSERT INTO notifications " +
				"(user_id, message, created_at) VALUES (?, ?, ?)",
		)

		if err == nil {
			stmt2.Exec(creditorID, message, now)
			stmt2.Close()
		}

		util.WriteJSON(w, map[string]string{
			"status":  "loan request created",
			"loan_id": loanID,
		})
	}
}

func LoanAccept(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		creditor, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			LoanID string `json:"loan_id"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var dbCreditorID int64
		err = db.QueryRow(
			"SELECT creditor_id FROM loans WHERE loan_id = ? AND status = 'pending'",
			req.LoanID,
		).Scan(&dbCreditorID)

		if err != nil {
			util.WriteJSONError(w, "Loan not found or not pending", http.StatusBadRequest)
			return
		}

		if dbCreditorID != creditor.ID {
			util.WriteJSONError(w, "Unauthorized: not the creditor", http.StatusUnauthorized)
			return
		}

		stmt, err := db.Prepare("UPDATE loans SET status = 'accepted' WHERE loan_id = ?")
		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(req.LoanID)
		if err != nil {
			util.WriteJSONError(w, "Error updating loan status", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{"status": "loan accepted"})
	}
}

func LoanReject(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		creditor, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			LoanID        string `json:"loan_id"`
			RejectionCode string `json:"rejection_code"`
			Message       string `json:"message"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var dbCreditorID int64
		err = db.QueryRow(
			"SELECT creditor_id FROM loans WHERE loan_id = ? AND status = 'pending'",
			req.LoanID,
		).Scan(&dbCreditorID)

		if err != nil {
			util.WriteJSONError(w, "Loan not found or not pending", http.StatusBadRequest)
			return
		}

		if dbCreditorID != creditor.ID {
			util.WriteJSONError(w, "Unauthorized: not the creditor", http.StatusUnauthorized)
			return
		}

		newStatus := "rejected:" + req.RejectionCode + ":" + req.Message
		stmt, err := db.Prepare("UPDATE loans SET status = ? WHERE loan_id = ?")

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(newStatus, req.LoanID)
		if err != nil {
			util.WriteJSONError(w, "Error updating loan status", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{"status": "loan rejected"})
	}
}

func PaymentTransaction(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		payer, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			RecipientIdentifier string  `json:"recipient_identifier"`
			Amount              float64 `json:"amount"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if payer.BalanceUra < req.Amount {
			util.WriteJSONError(w, "Insufficient funds", http.StatusBadRequest)
			return
		}
		twoBusinessDaysAgo := time.Now().Add(-48 * time.Hour)

		var receivedSum int64
		err = db.QueryRow(
			"SELECT COALESCE(SUM(amount),0) FROM transactions "+
				"WHERE user_id = ? AND created_at > ? AND category = 'incoming'",
			payer.ID, twoBusinessDaysAgo.Format(time.RFC3339),
		).Scan(&receivedSum)

		if err == nil && receivedSum >= 50000 {
			util.WriteJSONError(w, "Received funds limit exceeded in past 2 business days", http.StatusBadRequest)
			return
		}

		if req.Amount > 2000 {
			var totalLoans int64
			err = db.QueryRow(
				"SELECT COALESCE(SUM(amount),0) FROM loans "+
					"WHERE debtor_id = ? AND status = 'accepted' AND loan_type = 'uro'",
				payer.ID,
			).Scan(&totalLoans)

			if err == nil && totalLoans > 5000 {
				util.WriteJSONError(w, "Loan limit exceeded for payments over 2k", http.StatusBadRequest)
				return
			}
		}

		if req.Amount > 200000 {
			util.WriteJSONError(w, "Payment amount exceeds 200k uro", http.StatusBadRequest)
			return
		}

		var recipientID int64
		err = db.QueryRow(
			"SELECT id FROM users WHERE identifier = ?",
			req.RecipientIdentifier,
		).Scan(&recipientID)

		if err != nil {
			util.WriteJSONError(w, "Recipient not found", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(
			"UPDATE users SET balance_ura = balance_ura - ? WHERE id = ?",
			req.Amount,
			payer.ID,
		)

		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, "Error updating payer balance", http.StatusInternalServerError)

			return
		}

		_, err = tx.Exec(
			"UPDATE users SET balance_ura = balance_ura + ? WHERE id = ?",
			req.Amount,
			recipientID,
		)

		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, "Error updating recipient balance", http.StatusInternalServerError)

			return
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, "Error generating transaction ID", http.StatusInternalServerError)

			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		_, err = tx.Exec(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at) "+
				"VALUES (?, ?, 'outgoing', ?, ?)",
			transactionID,
			payer.ID,
			req.Amount,
			now,
		)

		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, "Error recording transaction", http.StatusInternalServerError)

			return
		}

		_, err = tx.Exec(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at) "+
				"VALUES (?, ?, 'incoming', ?, ?)",
			transactionID,
			recipientID,
			req.Amount,
			now,
		)

		if err != nil {
			tx.Rollback()
			util.WriteJSONError(w, "Error recording transaction", http.StatusInternalServerError)

			return
		}

		err = tx.Commit()
		if err != nil {
			util.WriteJSONError(w, "Error finalizing transaction", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":         "payment transaction successful",
			"transaction_id": transactionID,
		})
	}
}

func PaymentRequest(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			Amount float64 `json:"amount"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Amount > 100000 {
			util.WriteJSONError(w, "Payment request cannot be more than 100k uro", http.StatusBadRequest)
			return
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, "Error generating transaction ID", http.StatusInternalServerError)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt, err := db.Prepare(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at) " +
				"VALUES (?, ?, 'payment_request', ?, ?)",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(transactionID, user.ID, req.Amount, now)
		if err != nil {
			util.WriteJSONError(w, "Error recording payment request", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":         "payment request recorded",
			"transaction_id": transactionID,
		})
	}
}

func RefundRequest(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		_, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			LoanID string `json:"loan_id"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		refundID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, "Error generating refund ID", http.StatusInternalServerError)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt, err := db.Prepare(
			"INSERT INTO refunds (refund_id, loan_id, status, created_at) " +
				"VALUES (?, ?, 'pending', ?)",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(refundID, req.LoanID, now)
		if err != nil {
			util.WriteJSONError(w, "Error creating refund request", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":    "refund request created",
			"refund_id": refundID,
		})
	}
}

func RefundReject(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		_, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			RefundID string `json:"refund_id"`
			Message  string `json:"message"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare("UPDATE refunds SET status = ? WHERE refund_id = ?")
		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		newStatus := "rejected:" + req.Message
		_, err = stmt.Exec(newStatus, req.RefundID)

		if err != nil {
			util.WriteJSONError(w, "Error updating refund request", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status": "refund request rejected",
		})
	}
}

func RefundProcess(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		_, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			RefundID string `json:"refund_id"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare(
			"UPDATE refunds SET status = 'processed' WHERE refund_id = ?",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(req.RefundID)
		if err != nil {
			util.WriteJSONError(w, "Error processing refund", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{"status": "refund processed"})
	}
}

func Withdraw(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			Amount float64 `json:"amount"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if req.Amount >= 50000 {
			util.WriteJSONError(w, "Withdraw amount must be less than 50k uro", http.StatusBadRequest)
			return
		}
		twoBusinessDaysAgo := time.Now().Add(-48 * time.Hour)

		var receivedSum int64
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
				http.StatusBadRequest,
			)
			return
		}

		var balanceUra float64
		stmt, err := db.Query(
			"SELECT balance_ura FROM users WHERE id = " + strconv.Itoa(int(user.ID)),
		)
		stmt.Scan(&balanceUra)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}

		if balanceUra < req.Amount {
			util.WriteJSONError(w, "Insufficient funds", http.StatusBadRequest)
			return
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, "Error generating transaction ID", http.StatusInternalServerError)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt2, err := db.Prepare(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at, processed) " +
				"VALUES (?, ?, 'withdraw', ?, ?, false)",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}

		stmt2.Exec(transactionID, user.ID, req.Amount, now)
		stmt2.Close()

		util.WriteJSON(w, map[string]string{
			"status":         "200",
			"transaction_id": transactionID,
		})
	}
}

func CashIn(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		var req struct {
			Amount string `json:"amount"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		amount, err := strconv.ParseFloat(req.Amount, 64)
		if err != nil {
			util.WriteJSONError(w, "Invalid amount value", http.StatusBadRequest)
			return
		}

		if amount >= 100000 {
			util.WriteJSONError(w, "Cash in amount must be less than 100k uro", http.StatusBadRequest)
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
				util.WriteJSONError(w, "Cash in allowed only every 12 hours", http.StatusBadRequest)
				return
			}
		}

		transactionID, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, "Error generating transaction ID", http.StatusInternalServerError)
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		stmt2, err := db.Prepare(
			"INSERT INTO transactions (transaction_id, user_id, category, amount, created_at, processed) " +
				"VALUES (?, ?, 'cashin', ?, ?, false)",
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}

		stmt2.Exec(transactionID, user.ID, amount, now)
		stmt2.Close()

		util.WriteJSON(w, map[string]string{
			"status":         "200",
			"transaction_id": transactionID,
		})
	}
}

func UserLogin(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			util.WriteJSONError(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if !util.IsValidSHA512(req.Password) {
			util.WriteJSONError(w, "Password is not a valid SHA-512", http.StatusBadRequest)
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
			util.WriteJSONError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		sessionToken, err := util.GenerateRandomIdentifier(256)
		if err != nil {
			util.WriteJSONError(w, "Error generating session token", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().Add(timeoutMinute * time.Minute).UTC().Format(time.RFC3339)
		stmt, err := db.Prepare("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)")

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(sessionToken, user.ID, expiresAt)
		if err != nil {
			util.WriteJSONError(w, "Error creating session", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status":        "200",
			"session_token": sessionToken,
			"security_code": user.SecurityCode,
		})
	}
}

func UserFetchInfo(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		rows, err := db.Query(
			"SELECT transaction_id, category, amount, created_at "+
				"FROM transactions WHERE user_id = ?",
			user.ID,
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var transactions []map[string]interface{}
		for rows.Next() {
			var tid, category, createdAt string
			var amount float64

			rows.Scan(&tid, &category, &amount, &createdAt)
			transactions = append(transactions, map[string]interface{}{
				"transaction_id": tid,
				"category":       category,
				"amount":         amount,
				"created_at":     createdAt,
			})
		}

		rows2, err := db.Query(
			"SELECT loan_id, amount, loan_type, timespan, payment_type, status, created_at "+
				"FROM loans WHERE debtor_id = ? OR creditor_id = ?",
			user.ID,
			user.ID,
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows2.Close()

		var loans []map[string]interface{}
		for rows2.Next() {
			var loanID, loanType, paymentType, status, createdAt string
			var amount int64
			var timespan int

			rows2.Scan(
				&loanID,
				&amount,
				&loanType,
				&timespan,
				&paymentType,
				&status,
				&createdAt,
			)

			loans = append(loans, map[string]interface{}{
				"loan_id":      loanID,
				"amount":       amount,
				"loan_type":    loanType,
				"timespan":     timespan,
				"payment_type": paymentType,
				"status":       status,
				"created_at":   createdAt,
			})
		}

		util.WriteJSON(w, map[string]interface{}{
			"user":         user,
			"transactions": transactions,
			"loans":        loans,
		})
	}
}

func UserLogout(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionToken := r.Header.Get("X-Session-Token")
		if sessionToken == "" {
			util.WriteJSONError(w, "Missing session token", http.StatusUnauthorized)
			return
		}

		stmt, err := db.Prepare("DELETE FROM sessions WHERE token = ?")
		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}

		_, err = stmt.Exec(sessionToken)
		if err != nil {
			util.WriteJSONError(w, "Error logging out", http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, map[string]string{
			"status": "logout successful",
		})
	}
}

func UserFetchNotifications(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		user, err := authenticate(db, r)
		if err != nil {
			util.WriteJSONError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		rows, err := db.Query(
			"SELECT id, message, created_at, is_read "+
				"FROM notifications WHERE user_id = ? AND is_read = 0",
			user.ID,
		)

		if err != nil {
			util.WriteJSONError(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var notifications []map[string]interface{}
		for rows.Next() {
			var id int64
			var message, createdAt string
			var isRead int

			rows.Scan(&id, &message, &createdAt, &isRead)
			notifications = append(notifications, map[string]interface{}{
				"id":         id,
				"message":    message,
				"created_at": createdAt,
				"is_read":    isRead,
			})
		}

		stmt, err := db.Prepare("UPDATE notifications SET is_read = 1 WHERE user_id = ?")
		if err == nil {
			stmt.Exec(user.ID)
			stmt.Close()
		}

		util.WriteJSON(w, map[string]interface{}{
			"notifications": notifications,
		})
	}
}

func ValidateSession(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			util.WriteJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionToken := r.Header.Get("X-Session-Token")
		if sessionToken == "" {
			util.WriteJSONError(w, "Missing session token", http.StatusUnauthorized)
			return
		}

		var userID int64
		var expiresAtStr string

		err := db.QueryRow(
			"SELECT user_id, expires_at FROM sessions WHERE token = ?",
			sessionToken,
		).Scan(&userID, &expiresAtStr)

		if err != nil {
			util.WriteJSONError(w, "Invalid session token", http.StatusUnauthorized)
			return
		}

		expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			util.WriteJSONError(w, "Session time parse error", http.StatusInternalServerError)
			return
		}

		hasExpired := time.Now().After(expiresAt)
		if hasExpired {
			if _, err := db.Exec(
				"DELETE FROM sessions WHERE token = ?",
				sessionToken,
			); err != nil {
				util.WriteJSONError(w, "Internal error occured", http.StatusUnauthorized)
				return
			}
		}

		util.WriteJSON(w, map[string]interface{}{
			"status":     "200",
			"user_id":    strconv.FormatInt(userID, 10),
			"expired":    strconv.FormatBool(hasExpired),
			"expires_at": expiresAtStr,
		})
	}
}
