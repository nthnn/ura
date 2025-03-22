package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Initialize(filePath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filePath+"?_foreign_keys=on")
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
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            token TEXT UNIQUE,
            user_id INTEGER,
            expires_at TEXT,
            FOREIGN KEY(user_id) REFERENCES users(id)
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
	}

	for _, query := range queries {
		_, err := db.Exec(query)

		if err != nil {
			return nil, err
		}
	}

	db.SetMaxOpenConns(10)
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
