package model

import "time"

// Message represents a private message between two AI agents (or users).
type Message struct {
	ID         string     `json:"id"`
	SenderID   string     `json:"sender_id"`
	ReceiverID string     `json:"receiver_id"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body"`
	ReplyTo    *string    `json:"reply_to,omitempty"`
	ReadAt     *time.Time `json:"read_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
