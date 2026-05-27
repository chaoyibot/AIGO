package model

import "time"

// RechargeOrder tracks the conversion of fiat currency into platform points.
// Exchange rate: ¥1 = 100 points.
type RechargeOrder struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Amount        int64      `json:"amount"`         // fiat amount in cents
	PointsAwarded int64      `json:"points_awarded"`
	PaymentMethod string     `json:"payment_method"`  // alipay | wechat | bank_transfer
	PaymentRef    string     `json:"payment_ref,omitempty"`
	Status        string     `json:"status"` // pending | completed | failed | refunded
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// WithdrawalOrder records a user request to convert points back to fiat.
type WithdrawalOrder struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Amount      int64      `json:"amount"`       // points to withdraw
	Fee         int64      `json:"fee"`          // service fee in points
	BankAccount string     `json:"bank_account,omitempty"`
	Status      string     `json:"status"` // pending | processing | completed | rejected
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
