package repository

import (
	"database/sql"
	"fmt"
	"github.com/aigo/internal/model"
	"github.com/lib/pq"
)

type WebhookRepo struct {
	db *sql.DB
}

func NewWebhookRepo(db *sql.DB) *WebhookRepo {
	return &WebhookRepo{db: db}
}

func (r *WebhookRepo) Create(h *model.Webhook) error {
	_, err := r.db.Exec(
		`INSERT INTO webhooks (id, user_id, url, secret, events, active)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		h.ID, h.UserID, h.URL, h.Secret, pq.Array(h.Events), h.Active)
	return err
}

func (r *WebhookRepo) FindByID(id string) (*model.Webhook, error) {
	h := &model.Webhook{}
	err := r.db.QueryRow(
		`SELECT id, user_id, url, secret, events, active, created_at
		 FROM webhooks WHERE id = $1`, id,
	).Scan(&h.ID, &h.UserID, &h.URL, &h.Secret, pq.Array(&h.Events), &h.Active, &h.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find webhook: %w", err)
	}
	return h, nil
}

func (r *WebhookRepo) ListByUser(userID string) ([]model.Webhook, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, url, secret, events, active, created_at
		 FROM webhooks WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []model.Webhook
	for rows.Next() {
		var h model.Webhook
		if err := rows.Scan(&h.ID, &h.UserID, &h.URL, &h.Secret, pq.Array(&h.Events), &h.Active, &h.CreatedAt); err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}

func (r *WebhookRepo) ListByEvent(eventType string) ([]model.Webhook, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, url, secret, events, active, created_at
		 FROM webhooks WHERE $1 = ANY(events) AND active = true`, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []model.Webhook
	for rows.Next() {
		var h model.Webhook
		if err := rows.Scan(&h.ID, &h.UserID, &h.URL, &h.Secret, pq.Array(&h.Events), &h.Active, &h.CreatedAt); err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}

func (r *WebhookRepo) Update(h *model.Webhook) error {
	_, err := r.db.Exec(
		`UPDATE webhooks SET url=$1, secret=$2, events=$3, active=$4 WHERE id=$5 AND user_id=$6`,
		h.URL, h.Secret, pq.Array(h.Events), h.Active, h.ID, h.UserID)
	return err
}

func (r *WebhookRepo) Delete(id, userID string) error {
	_, err := r.db.Exec("DELETE FROM webhooks WHERE id=$1 AND user_id=$2", id, userID)
	return err
}

func (r *WebhookRepo) CreateLog(log *model.WebhookLog) error {
	_, err := r.db.Exec(
		`INSERT INTO webhook_logs (id, webhook_id, event_type, payload, response_code, response_body, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		log.ID, log.WebhookID, log.EventType, log.Payload, log.ResponseCode, log.ResponseBody, log.Status)
	return err
}

func (r *WebhookRepo) Logs(webhookID string, limit int) ([]model.WebhookLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.Query(
		`SELECT id, webhook_id, event_type, payload, response_code, response_body, status, next_retry_at, created_at
		 FROM webhook_logs WHERE webhook_id=$1 ORDER BY created_at DESC LIMIT $2`,
		webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.WebhookLog
	for rows.Next() {
		var l model.WebhookLog
		if err := rows.Scan(&l.ID, &l.WebhookID, &l.EventType, &l.Payload, &l.ResponseCode, &l.ResponseBody, &l.Status, &l.NextRetryAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}
