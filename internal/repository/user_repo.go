package repository

import (
	"database/sql"
	"fmt"

	"github.com/aigo/internal/model"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *model.User) error {
	_, err := r.db.Exec(
		"INSERT INTO users (id, public_key, nickname, role, status) VALUES ($1, $2, $3, $4, $5)",
		user.ID, user.PublicKey, user.Nickname, user.Role, user.Status,
	)
	return err
}

func (r *UserRepo) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		"SELECT id, public_key, nickname, role, status, created_at, updated_at FROM users WHERE id = $1", id,
	).Scan(&u.ID, &u.PublicKey, &u.Nickname, &u.Role, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) CreateAPIKey(key *model.APIKey) error {
	_, err := r.db.Exec(
		"INSERT INTO api_keys (id, user_id, key_hash, name, expires_at) VALUES ($1, $2, $3, $4, $5)",
		key.ID, key.UserID, key.KeyHash, key.Name, key.ExpiresAt,
	)
	return err
}

func (r *UserRepo) FindAPIKeyByHash(hash string) (*model.APIKey, error) {
	k := &model.APIKey{}
	err := r.db.QueryRow(
		"SELECT id, user_id, key_hash, name, last_used_at, expires_at, created_at FROM api_keys WHERE key_hash = $1", hash,
	).Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Name, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find api key: %w", err)
	}
	return k, nil
}

func (r *UserRepo) UpdateAPIKeyLastUsed(id string) error {
	_, err := r.db.Exec("UPDATE api_keys SET last_used_at = now() WHERE id = $1", id)
	return err
}

func (r *UserRepo) ListAPIKeys(userID string) ([]model.APIKey, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, name, last_used_at, expires_at, created_at FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC", userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *UserRepo) DeleteAPIKey(id string) error {
	_, err := r.db.Exec("DELETE FROM api_keys WHERE id = $1", id)
	return err
}

func (r *UserRepo) UpdateNickname(userID, nickname string) error {
	_, err := r.db.Exec("UPDATE users SET nickname = $1, updated_at = now() WHERE id = $2", nickname, userID)
	return err
}
