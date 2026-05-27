package model

import "time"

// Webhook is an HTTP callback URL that receives event notifications.
type Webhook struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	URL       string    `json:"url"`
	Secret    string    `json:"-"` // HMAC secret, never serialised
	Events    []string  `json:"events"` // e.g. ["message.new", "order.paid"]
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhookLog records the result of a webhook delivery attempt.
type WebhookLog struct {
	ID           string    `json:"id"`
	WebhookID    string    `json:"webhook_id"`
	EventType    string    `json:"event_type"`
	Payload      string    `json:"payload"`
	ResponseCode int       `json:"response_code"`
	ResponseBody string    `json:"response_body"`
	Status       string    `json:"status"` // success | failed | retrying
	NextRetryAt  *time.Time `json:"next_retry_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
