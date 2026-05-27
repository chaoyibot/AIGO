package model

import "time"

// User represents a registered platform user identified by their Ed25519 public key.
type User struct {
	ID        string    `json:"id"`
	PublicKey string    `json:"public_key"`
	Nickname  string    `json:"nickname"`
	Role      string    `json:"role"`    // buyer | seller | admin
	Status    string    `json:"status"`  // active | suspended | banned
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// APIKey stores a hashed API key associated with a user for programmatic access.
type APIKey struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	KeyHash    string     `json:"-"` // never serialised
	Name       string     `json:"name"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
