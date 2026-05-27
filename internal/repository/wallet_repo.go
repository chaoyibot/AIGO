package repository

import (
	"database/sql"
	"fmt"
	"github.com/aigo/internal/model"
)

type WalletRepo struct {
	db *sql.DB
}

func NewWalletRepo(db *sql.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) Create(w *model.Wallet) error {
	_, err := r.db.Exec(
		"INSERT INTO wallets (id, user_id, fiat_balance, points_balance, currency) VALUES ($1, $2, $3, $4, $5)",
		w.ID, w.UserID, w.FiatBalance, w.PointsBalance, w.Currency)
	return err
}

func (r *WalletRepo) FindByUserID(userID string) (*model.Wallet, error) {
	w := &model.Wallet{}
	err := r.db.QueryRow(
		"SELECT id, user_id, fiat_balance, points_balance, currency, created_at, updated_at FROM wallets WHERE user_id = $1", userID,
	).Scan(&w.ID, &w.UserID, &w.FiatBalance, &w.PointsBalance, &w.Currency, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find wallet: %w", err)
	}
	return w, nil
}

func (r *WalletRepo) AddPoints(userID string, amount int64) error {
	_, err := r.db.Exec("UPDATE wallets SET points_balance = points_balance + $1, updated_at = now() WHERE user_id = $2", amount, userID)
	return err
}

func (r *WalletRepo) DeductPoints(userID string, amount int64) error {
	_, err := r.db.Exec("UPDATE wallets SET points_balance = points_balance - $1, updated_at = now() WHERE user_id = $2 AND points_balance >= $1", amount, userID)
	return err
}

func (r *WalletRepo) AddFiat(userID string, amount int64) error {
	_, err := r.db.Exec("UPDATE wallets SET fiat_balance = fiat_balance + $1, updated_at = now() WHERE user_id = $2", amount, userID)
	return err
}

func (r *WalletRepo) DeductFiat(userID string, amount int64) error {
	_, err := r.db.Exec("UPDATE wallets SET fiat_balance = fiat_balance - $1, updated_at = now() WHERE user_id = $2 AND fiat_balance >= $1", amount, userID)
	return err
}

// WithRowLock locks the wallet row for update (must be called within a transaction)
func (r *WalletRepo) FindByUserIDForUpdate(tx *sql.Tx, userID string) (*model.Wallet, error) {
	w := &model.Wallet{}
	err := tx.QueryRow(
		"SELECT id, user_id, fiat_balance, points_balance, currency FROM wallets WHERE user_id = $1 FOR UPDATE", userID,
	).Scan(&w.ID, &w.UserID, &w.FiatBalance, &w.PointsBalance, &w.Currency)
	if err != nil {
		return nil, fmt.Errorf("find wallet for update: %w", err)
	}
	return w, nil
}

// --- Transactions ---

func (r *WalletRepo) AddTransaction(tx *sql.Tx, t *model.Transaction) error {
	_, err := tx.Exec(
		"INSERT INTO transactions (id, user_id, type, amount, balance, reference) VALUES ($1, $2, $3, $4, $5, $6)",
		t.ID, t.UserID, t.Type, t.Amount, t.Balance, t.Reference)
	return err
}

func (r *WalletRepo) ListTransactions(userID string, limit int) ([]model.Transaction, error) {
	if limit <= 0 || limit > 100 { limit = 20 }
	rows, err := r.db.Query(
		"SELECT id, user_id, type, amount, balance, reference, created_at FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2",
		userID, limit)
	if err != nil { return nil, err }
	defer rows.Close()

	var txns []model.Transaction
	for rows.Next() {
		var t model.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.Balance, &t.Reference, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, nil
}

// --- Recharge Orders ---

type RechargeRepo struct {
	db *sql.DB
}

func NewRechargeRepo(db *sql.DB) *RechargeRepo {
	return &RechargeRepo{db: db}
}

func (r *RechargeRepo) Create(o *model.RechargeOrder) error {
	_, err := r.db.Exec(
		"INSERT INTO recharge_orders (id, user_id, amount, points_awarded, payment_method, status) VALUES ($1, $2, $3, $4, $5, $6)",
		o.ID, o.UserID, o.Amount, o.PointsAwarded, o.PaymentMethod, o.Status)
	return err
}

func (r *RechargeRepo) Complete(id string) error {
	_, err := r.db.Exec("UPDATE recharge_orders SET status='completed', completed_at=now() WHERE id=$1", id)
	return err
}

func (r *RechargeRepo) FindByID(id string) (*model.RechargeOrder, error) {
	o := &model.RechargeOrder{}
	err := r.db.QueryRow(
		"SELECT id, user_id, amount, points_awarded, payment_method, payment_ref, status, created_at, completed_at FROM recharge_orders WHERE id=$1", id,
	).Scan(&o.ID, &o.UserID, &o.Amount, &o.PointsAwarded, &o.PaymentMethod, &o.PaymentRef, &o.Status, &o.CreatedAt, &o.CompletedAt)
	if err != nil { return nil, fmt.Errorf("find recharge: %w", err) }
	return o, nil
}

// --- Withdrawal Orders ---

type WithdrawalRepo struct {
	db *sql.DB
}

func NewWithdrawalRepo(db *sql.DB) *WithdrawalRepo {
	return &WithdrawalRepo{db: db}
}

func (r *WithdrawalRepo) Create(o *model.WithdrawalOrder) error {
	_, err := r.db.Exec(
		"INSERT INTO withdrawal_orders (id, user_id, amount, fee, bank_account, status) VALUES ($1, $2, $3, $4, $5, $6)",
		o.ID, o.UserID, o.Amount, o.Fee, o.BankAccount, o.Status)
	return err
}

func (r *WithdrawalRepo) ListByUser(userID string) ([]model.WithdrawalOrder, error) {
	rows, err := r.db.Query("SELECT id, user_id, amount, fee, status, created_at, completed_at FROM withdrawal_orders WHERE user_id=$1 ORDER BY created_at DESC", userID)
	if err != nil { return nil, err }
	defer rows.Close()
	var orders []model.WithdrawalOrder
	for rows.Next() {
		var o model.WithdrawalOrder
		if err := rows.Scan(&o.ID, &o.UserID, &o.Amount, &o.Fee, &o.Status, &o.CreatedAt, &o.CompletedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}
