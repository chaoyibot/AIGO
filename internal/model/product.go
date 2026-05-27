package model

import "time"

// Product stores encrypted commodity metadata. Only the seller (and, via key
// sharing, the buyer after purchase) can decrypt the title, description, and
// arbitrary metadata fields.
//
// For normal product browsing (GET /products/:id), the API layer transparently
// decrypts these fields so other users can read the content. The encrypted
// fields are still stored on disk.
type Product struct {
	ID         string    `json:"id"`
	SellerID   string    `json:"seller_id"`
	PriceMin   int64     `json:"price_min"`
	PriceMax   int64     `json:"price_max"`
	Category   string    `json:"category"`
	Tags       []string  `json:"tags"`
	Status     string    `json:"status"` // draft | active | archived
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Encrypted blobs — stored encrypted on disk, never serialised to JSON.
	EncryptedTitle    []byte `json:"-"` // never serialised
	EncryptedDesc     []byte `json:"-"` // never serialised
	EncryptedMetadata []byte `json:"-"` // never serialised
	EncryptedKeySeller []byte `json:"-"` // encrypted symmetric key (seller copy)

	// Decrypted content — populated by the service layer when reading.
	// These fields ARE serialised to JSON in API responses so that other
	// users can view product details without needing any key.
	DecryptedTitle    string `json:"title,omitempty"`
	DecryptedDesc     string `json:"description,omitempty"`
	DecryptedMetadata string `json:"metadata,omitempty"`
}
