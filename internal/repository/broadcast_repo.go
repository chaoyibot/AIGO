package repository

import (
	"database/sql"
	"fmt"
	"github.com/aigo/internal/model"
)

type BroadcastRepo struct {
	db *sql.DB
}

func NewBroadcastRepo(db *sql.DB) *BroadcastRepo {
	return &BroadcastRepo{db: db}
}

func (r *BroadcastRepo) Create(b *model.Broadcast) error {
	_, err := r.db.Exec(
		`INSERT INTO broadcasts (id, type, title, content, sender_id, points_cost, link_url, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		b.ID, b.Type, b.Title, b.Content, b.SenderID, b.PointsCost, b.LinkURL, b.Status, b.ExpiresAt)
	return err
}

func (r *BroadcastRepo) ListActive() ([]model.Broadcast, error) {
	rows, err := r.db.Query(
		`SELECT id, type, title, content, sender_id, points_cost, link_url, created_at
		 FROM broadcasts WHERE status = 'active' ORDER BY created_at DESC LIMIT 50`)
	if err != nil { return nil, err }
	defer rows.Close()
	var broadcasts []model.Broadcast
	for rows.Next() {
		var b model.Broadcast
		if err := rows.Scan(&b.ID, &b.Type, &b.Title, &b.Content, &b.SenderID, &b.PointsCost, &b.LinkURL, &b.CreatedAt); err != nil {
			return nil, err
		}
		broadcasts = append(broadcasts, b)
	}
	return broadcasts, nil
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
