package repository

import (
	"database/sql"
	"fmt"
	"github.com/aigo/internal/model"
)

type MessageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Send(msg *model.Message) error {
	_, err := r.db.Exec(
		`INSERT INTO messages (id, sender_id, receiver_id, subject, body, reply_to)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		msg.ID, msg.SenderID, msg.ReceiverID, msg.Subject, msg.Body, msg.ReplyTo)
	return err
}

func (r *MessageRepo) Inbox(userID string, limit int) ([]model.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.Query(
		`SELECT id, sender_id, receiver_id, subject, body, reply_to, read_at, created_at
		 FROM messages WHERE receiver_id = $1
		 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		var replyTo sql.NullString
		var readAt sql.NullTime
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Subject, &m.Body, &replyTo, &readAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		if replyTo.Valid {
			m.ReplyTo = &replyTo.String
		}
		if readAt.Valid {
			m.ReadAt = &readAt.Time
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *MessageRepo) Sent(userID string, limit int) ([]model.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.Query(
		`SELECT id, sender_id, receiver_id, subject, body, reply_to, read_at, created_at
		 FROM messages WHERE sender_id = $1
		 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []model.Message
	for rows.Next() {
		var m model.Message
		var replyTo sql.NullString
		var readAt sql.NullTime
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Subject, &m.Body, &replyTo, &readAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		if replyTo.Valid {
			m.ReplyTo = &replyTo.String
		}
		if readAt.Valid {
			m.ReadAt = &readAt.Time
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *MessageRepo) UnreadCount(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE receiver_id = $1 AND read_at IS NULL", userID).Scan(&count)
	return count, err
}

func (r *MessageRepo) MarkRead(msgID, userID string) error {
	_, err := r.db.Exec(
		"UPDATE messages SET read_at = now() WHERE id = $1 AND receiver_id = $2 AND read_at IS NULL",
		msgID, userID)
	return err
}

func (r *MessageRepo) FindByID(id string) (*model.Message, error) {
	m := &model.Message{}
	var replyTo sql.NullString
	var readAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id, sender_id, receiver_id, subject, body, reply_to, read_at, created_at
		 FROM messages WHERE id = $1`, id,
	).Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.Subject, &m.Body, &replyTo, &readAt, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find message: %w", err)
	}
	if replyTo.Valid {
		m.ReplyTo = &replyTo.String
	}
	if readAt.Valid {
		m.ReadAt = &readAt.Time
	}
	return m, nil
}
