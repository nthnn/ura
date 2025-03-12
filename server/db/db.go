package db

import (
	"database/sql"
)

func Initialize() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "ura.s3db")
	if err != nil {
		return nil, err
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL,
            email TEXT NOT NULL,
            password TEXT NOT NULL,
            identifier TEXT NOT NULL,
            security_code TEXT NOT NULL,
            balance_ura REAL DEFAULT 0,
            created_at TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS sessions (
            token TEXT PRIMARY KEY,
            user_id INTEGER,
            expires_at TEXT,
            FOREIGN KEY(user_id) REFERENCES users(id)
        );`,
		`CREATE TABLE IF NOT EXISTS loans (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            loan_id TEXT,
            debtor_id INTEGER,
            creditor_id INTEGER,
            amount REAL,
            loan_type TEXT,
            timespan INTEGER,
            payment_type TEXT,
            status TEXT,
            created_at TEXT
        );`,
		`CREATE TABLE IF NOT EXISTS transactions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            transaction_id TEXT,
            user_id INTEGER,
            category TEXT,
            amount REAL,
            created_at TEXT,
            processed INTEGER DEFAULT 1
        );`,
		`CREATE TABLE IF NOT EXISTS notifications (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER,
            message TEXT,
            created_at TEXT,
            is_read INTEGER DEFAULT 0
        );`,
		`CREATE TABLE IF NOT EXISTS refunds (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            refund_id TEXT,
            loan_id TEXT,
            status TEXT,
            created_at TEXT
        );`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)

		if err != nil {
			return nil, err
		}
	}

	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(0)

	return db, nil
}
