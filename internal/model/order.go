package model

import "time"

// Order captures the lifecycle of a purchase from creation through escrow
// settlement or dispute resolution.
type Order struct {
	ID              string     `json:"id"`
	ListingID       string     `json:"listing_id"`
	BuyerID         string     `json:"buyer_id"`
	SellerID        string     `json:"seller_id"`
	ProductSnapshot string     `json:"product_snapshot,omitempty"`
	Quantity        int        `json:"quantity"`
	UnitPrice       int64      `json:"unit_price"`
	TotalPrice      int64      `json:"total_price"`
	Status          string     `json:"status"`       // pending | paid | confirmed | completed | disputed | cancelled
	EscrowStatus    string     `json:"escrow_status"` // pending | held | released | refunded
	DisputeReason   string     `json:"dispute_reason,omitempty"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Escrow holds funds in trust while an order is being fulfilled.
type Escrow struct {
	ID            string     `json:"id"`
	OrderID       string     `json:"order_id"`
	Amount        int64      `json:"amount"`
	Status        string     `json:"status"` // held | released | refunded | disputed
	AutoReleaseAt *time.Time `json:"auto_release_at,omitempty"`
	ReleasedAt    *time.Time `json:"released_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
