package wallet

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/aigo/internal/config"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type Service struct {
	walletRepo   *repository.WalletRepo
	rechargeRepo *repository.RechargeRepo
	withdrawRepo *repository.WithdrawalRepo
	cfg          *config.Config
	db           *sql.DB
}

func NewService(walletRepo *repository.WalletRepo, rechargeRepo *repository.RechargeRepo, withdrawRepo *repository.WithdrawalRepo, cfg *config.Config, db *sql.DB) *Service {
	return &Service{
		walletRepo:   walletRepo,
		rechargeRepo: rechargeRepo,
		withdrawRepo: withdrawRepo,
		cfg:          cfg,
		db:           db,
	}
}

func (s *Service) GetOrCreateWallet(userID string) (*model.Wallet, error) {
	w, err := s.walletRepo.FindByUserID(userID)
	if err == nil {
		return w, nil
	}
	w = &model.Wallet{
		ID:            uuid.New().String(),
		UserID:        userID,
		FiatBalance:   0,
		PointsBalance: 0,
		Currency:      "CNY",
	}
	if err := s.walletRepo.Create(w); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return w, nil
}

func (s *Service) GetBalance(userID string) (*model.Wallet, error) {
	return s.GetOrCreateWallet(userID)
}

// Recharge fiat money, returns points awarded
func (s *Service) Recharge(userID string, amount int64) (*model.RechargeOrder, error) {
	rate := s.cfg.Points.ExchangeRate // 1 fiat cent = rate points
	points := amount * rate

	order := &model.RechargeOrder{
		ID:            uuid.New().String(),
		UserID:        userID,
		Amount:        amount,
		PointsAwarded: points,
		PaymentMethod: "simulated",
		Status:        "completed",
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create wallet if not exists
	_, err = s.walletRepo.FindByUserIDForUpdate(tx, userID)
	if err != nil {
		w := &model.Wallet{ID: uuid.New().String(), UserID: userID, FiatBalance: 0, PointsBalance: 0, Currency: "CNY"}
		if err := s.walletRepo.Create(w); err != nil {
			return nil, err
		}
	}

	// Add fiat and points
	if _, err := tx.Exec("UPDATE wallets SET fiat_balance = fiat_balance + $1, points_balance = points_balance + $2, updated_at = now() WHERE user_id = $3", amount, points, userID); err != nil {
		return nil, err
	}

	// Create recharge order
	if _, err := tx.Exec("INSERT INTO recharge_orders (id, user_id, amount, points_awarded, payment_method, status, completed_at) VALUES ($1, $2, $3, $4, $5, 'completed', now())", order.ID, userID, amount, points, "simulated"); err != nil {
		return nil, err
	}

	// Add transaction record
	wallet, _ := s.walletRepo.FindByUserID(userID)
	if _, err := tx.Exec("INSERT INTO transactions (id, user_id, type, amount, balance, reference) VALUES ($1, $2, 'recharge', $3, $4, $5)", uuid.New().String(), userID, amount, wallet.FiatBalance, "充值"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("recharge commit: %w", err)
	}

	order.Status = "completed"
	return order, nil
}

// Convert fiat to points
func (s *Service) ConvertToPoints(userID string, fiatAmount int64) error {
	rate := s.cfg.Points.ExchangeRate
	points := fiatAmount * rate

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	wallet, err := s.walletRepo.FindByUserIDForUpdate(tx, userID)
	if err != nil {
		return fmt.Errorf("wallet not found")
	}

	if wallet.FiatBalance < fiatAmount {
		return fmt.Errorf("insufficient fiat balance")
	}

	newFiat := wallet.FiatBalance - fiatAmount
	newPoints := wallet.PointsBalance + points

	if _, err := tx.Exec("UPDATE wallets SET fiat_balance=$1, points_balance=$2, updated_at=now() WHERE user_id=$3", newFiat, newPoints, userID); err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT INTO transactions (id, user_id, type, amount, balance, reference) VALUES ($1, $2, 'conversion', $3, $4, '法币转积分')",
		uuid.New().String(), userID, points, newPoints); err != nil {
		return err
	}

	return tx.Commit()
}

// Get transaction history
func (s *Service) GetTransactions(userID string, limit int) ([]model.Transaction, error) {
	return s.walletRepo.ListTransactions(userID, limit)
}
