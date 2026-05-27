package broadcast

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/aigo/internal/config"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type Service struct {
	broadcastRepo *repository.BroadcastRepo
	walletRepo    *repository.WalletRepo
	cfg           *config.Config
	db            *sql.DB
}

func NewService(broadcastRepo *repository.BroadcastRepo, walletRepo *repository.WalletRepo, cfg *config.Config, db *sql.DB) *Service {
	return &Service{broadcastRepo: broadcastRepo, walletRepo: walletRepo, cfg: cfg, db: db}
}

// PublishSystem creates a free system broadcast
func (s *Service) PublishSystem(title, content string) (*model.Broadcast, error) {
	b := &model.Broadcast{
		ID:      uuid.New().String(),
		Type:    "system",
		Title:   title,
		Content: content,
		Status:  "active",
	}
	if err := s.broadcastRepo.Create(b); err != nil {
		return nil, fmt.Errorf("create system broadcast: %w", err)
	}
	return b, nil
}

// PublishCommercial creates a paid commercial broadcast, deducting points
func (s *Service) PublishCommercial(senderID, title, content, level, linkURL string) (*model.Broadcast, error) {
	costs := map[string]int64{
		"basic":   s.cfg.Broadcast.BasicCost,
		"standard": s.cfg.Broadcast.StandardCost,
		"premium":  s.cfg.Broadcast.PremiumCost,
	}
	cost, ok := costs[level]
	if !ok {
		return nil, fmt.Errorf("invalid level: %s (basic/standard/premium)", level)
	}

	tx, err := s.db.Begin()
	if err != nil { return nil, err }
	defer tx.Rollback()

	// Lock wallet
	w, err := s.walletRepo.FindByUserIDForUpdate(tx, senderID)
	if err != nil { return nil, fmt.Errorf("wallet not found") }
	if w.PointsBalance < cost {
		return nil, fmt.Errorf("insufficient points: have %d, need %d", w.PointsBalance, cost)
	}

	// Deduct points
	if _, err := tx.Exec("UPDATE wallets SET points_balance = points_balance - $1, updated_at = now() WHERE user_id = $2", cost, senderID); err != nil {
		return nil, err
	}

	// Create broadcast
	b := &model.Broadcast{
		ID:         uuid.New().String(),
		Type:       "commercial",
		Title:      title,
		Content:    content,
		SenderID:   &senderID,
		PointsCost: cost,
		LinkURL:    linkURL,
		Status:     "active",
	}
	if _, err := tx.Exec(
		`INSERT INTO broadcasts (id, type, title, content, sender_id, points_cost, link_url, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		b.ID, b.Type, b.Title, b.Content, b.SenderID, b.PointsCost, b.LinkURL, b.Status); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commercial broadcast commit: %w", err)
	}
	return b, nil
}

func (s *Service) ListBroadcasts() ([]model.Broadcast, error) {
	return s.broadcastRepo.ListActive()
}

func (s *Service) MarkRead(broadcastID, userID string) error {
	return s.broadcastRepo.MarkRead(broadcastID, userID)
}

func (s *Service) UnreadCount(userID string) (int, error) {
	return s.broadcastRepo.UnreadCount(userID)
}
