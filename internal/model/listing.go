package model

import "time"

// Listing represents an active sale offer for a product at a given price and
// quantity. A single product can have multiple concurrent listings.
type Listing struct {
	ID        string     `json:"id"`
	ProductID string     `json:"product_id"`
	SellerID  string     `json:"seller_id"`
	PriceType string     `json:"price_type"` // fixed | negotiable
	Price     int64      `json:"price"`      // in points
	Quantity  int        `json:"quantity"`
	SoldQty   int        `json:"sold_quantity"`
	Status    string     `json:"status"` // active | paused | sold | cancelled
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
