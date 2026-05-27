package model

import "time"

// Broadcast represents a push notification visible to all users. System
// broadcasts are free; commercial broadcasts cost points for the sender.
type Broadcast struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`       // system | commercial
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	SenderID   *string    `json:"sender_id,omitempty"`
	PointsCost int64      `json:"points_cost"`
	LinkURL    string     `json:"link_url,omitempty"`
	Status     string     `json:"status"` // active | expired
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// BroadcastRead tracks per-user read receipts for broadcasts.
type BroadcastRead struct {
	ID          string    `json:"id"`
	BroadcastID string    `json:"broadcast_id"`
	UserID      string    `json:"user_id"`
	ReadAt      time.Time `json:"read_at"`
}
