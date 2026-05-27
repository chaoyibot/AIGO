package model

import "time"

// Wallet tracks both fiat (in CNY cents) and points balances for a user.
// Points are the internal trading currency (¥1 = 100 points).
type Wallet struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	FiatBalance   int64     `json:"fiat_balance"`   // in cents (CNY)
	PointsBalance int64     `json:"points_balance"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Transaction records every balance-affecting operation on a wallet.
type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // recharge | conversion | purchase | broadcast | withdrawal
	Amount    int64     `json:"amount"`
	Balance   int64     `json:"balance"`            // running balance after this tx
	Reference string    `json:"reference,omitempty"` // external ref (order ID, etc.)
	CreatedAt time.Time `json:"created_at"`
}
