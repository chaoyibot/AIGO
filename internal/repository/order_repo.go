package repository

import (
	"database/sql"
	"fmt"
	"github.com/aigo/internal/model"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) FindByID(id string) (*model.Order, error) {
	o := &model.Order{}
	err := r.db.QueryRow(
		`SELECT id, listing_id, buyer_id, seller_id, quantity, unit_price, total_price, status, escrow_status, dispute_reason, confirmed_at, created_at, updated_at
		 FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.ListingID, &o.BuyerID, &o.SellerID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.EscrowStatus, &o.DisputeReason, &o.ConfirmedAt, &o.CreatedAt, &o.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("find order: %w", err) }
	return o, nil
}

func (r *OrderRepo) ListByUser(userID string, limit int) ([]model.Order, error) {
	if limit <= 0 || limit > 100 { limit = 20 }
	rows, err := r.db.Query(
		`SELECT id, listing_id, buyer_id, seller_id, quantity, unit_price, total_price, status, escrow_status, created_at
		 FROM orders WHERE buyer_id = $1 OR seller_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	var orders []model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.ListingID, &o.BuyerID, &o.SellerID, &o.Quantity, &o.UnitPrice, &o.TotalPrice, &o.Status, &o.EscrowStatus, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepo) Confirm(id string) error {
	_, err := r.db.Exec("UPDATE orders SET status='completed', confirmed_at=now(), updated_at=now() WHERE id=$1", id)
	return err
}

func (r *OrderRepo) Cancel(id string) error {
	_, err := r.db.Exec("UPDATE orders SET status='cancelled', updated_at=now() WHERE id=$1", id)
	return err
}

func (r *OrderRepo) Dispute(id, reason string) error {
	_, err := r.db.Exec("UPDATE orders SET status='disputed', dispute_reason=$1, updated_at=now() WHERE id=$2", reason, id)
	return err
}
