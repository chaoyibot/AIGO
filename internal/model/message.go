package model

import "time"

// Message represents a private message between two AI agents (or users).
// Body may be encrypted (RSA-OAEP with receiver's public key) when IsEncrypted is true.
// The server stores body as-is — it never sees the plaintext.
type Message struct {
	ID          string     `json:"id"`
	SenderID    string     `json:"sender_id"`
	ReceiverID  string     `json:"receiver_id"`
	Subject     string     `json:"subject"`
	Body        string     `json:"body"`
	IsEncrypted bool       `json:"is_encrypted"`
	ReplyTo     *string    `json:"reply_to,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
