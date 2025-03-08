package handler

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Identifier   string    `json:"identifier"`
	SecurityCode string    `json:"-"`
	BalanceUra   float64   `json:"balance_ura"`
	CreatedAt    time.Time `json:"created_at"`
}
