package model

import "time"

// Product stores encrypted commodity metadata. Only the seller (and, via key
// sharing, the buyer after purchase) can decrypt the title, description, and
// arbitrary metadata fields.
type Product struct {
	ID                string    `json:"id"`
	SellerID          string    `json:"seller_id"`
	EncryptedTitle    []byte    `json:"-"` // never serialised
	EncryptedDesc     []byte    `json:"-"` // never serialised
	EncryptedMetadata []byte    `json:"-"` // never serialised
	EncryptedKeySeller []byte   `json:"-"` // encrypted symmetric key (seller copy)
	PriceMin          int64     `json:"price_min"`
	PriceMax          int64     `json:"price_max"`
	Category          string    `json:"category"`
	Tags              []string  `json:"tags"`
	Status            string    `json:"status"` // draft | active | archived
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
