package repository

import (
	"database/sql"
	"time"

	"github.com/aigo/internal/model"
)

type BroadcastRepo struct {
	db *sql.DB
}

func NewBroadcastRepo(db *sql.DB) *BroadcastRepo {
	return &BroadcastRepo{db: db}
}

// Create inserts a broadcast with encrypted title/content already provided as bytea.
func (r *BroadcastRepo) Create(b *model.Broadcast) error {
	_, err := r.db.Exec(
		`INSERT INTO broadcasts (id, type, sender_id, points_cost, link_url, status, is_pinned, pinned_at, pin_expires_at, expires_at, encrypted_title, encrypted_content, immutable)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		b.ID, b.Type, b.SenderID, b.PointsCost, b.LinkURL, b.Status, b.IsPinned, b.PinnedAt, b.PinExpiresAt, b.ExpiresAt, b.EncryptedTitle, b.EncryptedContent, b.Immutable)
	return err
}

// ListActive returns active broadcasts with pinned ones surfaced first,
// then sorted by created_at DESC. Expired pinned broadcasts are unpinned automatically.
// Returns EncryptedTitle/EncryptedContent as bytea — the service layer decrypts them.
func (r *BroadcastRepo) ListActive() ([]model.Broadcast, error) {
	// First, auto-unpin any expired pinned broadcasts
	_, _ = r.db.Exec(`UPDATE broadcasts SET is_pinned = false WHERE is_pinned = true AND pin_expires_at IS NOT NULL AND pin_expires_at < now()`)

	rows, err := r.db.Query(
		`SELECT id, type, sender_id, points_cost, link_url, status, is_pinned, pinned_at, pin_expires_at, expires_at, created_at, immutable, encrypted_title, encrypted_content
		 FROM broadcasts
		 WHERE status = 'active'
		 ORDER BY is_pinned DESC, pin_expires_at DESC NULLS LAST, created_at DESC
		 LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var broadcasts []model.Broadcast
	for rows.Next() {
		var b model.Broadcast
		var encryptedTitle, encryptedContent []byte
		if err := rows.Scan(&b.ID, &b.Type, &b.SenderID, &b.PointsCost, &b.LinkURL, &b.Status,
			&b.IsPinned, &b.PinnedAt, &b.PinExpiresAt, &b.ExpiresAt, &b.CreatedAt, &b.Immutable,
			&encryptedTitle, &encryptedContent); err != nil {
			return nil, err
		}
		b.EncryptedTitle = encryptedTitle
		b.EncryptedContent = encryptedContent
		broadcasts = append(broadcasts, b)
	}
	return broadcasts, nil
}

// GetByID returns a single broadcast by ID with encrypted blobs.
func (r *BroadcastRepo) GetByID(id string) (*model.Broadcast, error) {
	var b model.Broadcast
	var encryptedTitle, encryptedContent []byte
	err := r.db.QueryRow(
		`SELECT id, type, sender_id, points_cost, link_url, status, is_pinned, pinned_at, pin_expires_at, expires_at, created_at, immutable, encrypted_title, encrypted_content
		 FROM broadcasts WHERE id = $1`, id).Scan(
		&b.ID, &b.Type, &b.SenderID, &b.PointsCost, &b.LinkURL, &b.Status,
		&b.IsPinned, &b.PinnedAt, &b.PinExpiresAt, &b.ExpiresAt, &b.CreatedAt, &b.Immutable,
		&encryptedTitle, &encryptedContent)
	if err != nil {
		return nil, err
	}
	b.EncryptedTitle = encryptedTitle
	b.EncryptedContent = encryptedContent
	return &b, nil
}

func (r *BroadcastRepo) SetPin(id string, pinExpiresAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE broadcasts SET is_pinned = true, pinned_at = now(), pin_expires_at = $2 WHERE id = $1`,
		id, pinExpiresAt)
	return err
}

func (r *BroadcastRepo) Unpin(id string) error {
	_, err := r.db.Exec(
		`UPDATE broadcasts SET is_pinned = false, pinned_at = NULL, pin_expires_at = NULL WHERE id = $1`,
		id)
	return err
}

func (r *BroadcastRepo) MarkRead(broadcastID, userID string) error {
	_, err := r.db.Exec(
		"INSERT INTO broadcast_reads (broadcast_id, user_id) VALUES ($1, $2) ON CONFLICT (broadcast_id, user_id) DO NOTHING",
		broadcastID, userID)
	return err
}

func (r *BroadcastRepo) UnreadCount(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM broadcasts b
		 WHERE b.status = 'active'
		 AND NOT EXISTS (SELECT 1 FROM broadcast_reads br WHERE br.broadcast_id = b.id AND br.user_id = $1)`, userID).Scan(&count)
	return count, err
}