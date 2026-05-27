package repository

import (
	"database/sql"
	"fmt"

	"github.com/aigo/internal/model"
)

type ListingRepo struct {
	db *sql.DB
}

func NewListingRepo(db *sql.DB) *ListingRepo {
	return &ListingRepo{db: db}
}

func (r *ListingRepo) Create(l *model.Listing) error {
	_, err := r.db.Exec(
		`INSERT INTO listings (id, product_id, seller_id, price_type, price, quantity, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		l.ID, l.ProductID, l.SellerID, l.PriceType, l.Price, l.Quantity, l.Status, l.ExpiresAt)
	return err
}

func (r *ListingRepo) FindByID(id string) (*model.Listing, error) {
	l := &model.Listing{}
	err := r.db.QueryRow(
		`SELECT id, product_id, seller_id, price_type, price, quantity, sold_quantity, status, expires_at, created_at
		 FROM listings WHERE id = $1`, id,
	).Scan(&l.ID, &l.ProductID, &l.SellerID, &l.PriceType, &l.Price, &l.Quantity, &l.SoldQty, &l.Status, &l.ExpiresAt, &l.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find listing: %w", err)
	}
	return l, nil
}

func (r *ListingRepo) ListActive(limit int) ([]model.Listing, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.Query(
		`SELECT id, product_id, seller_id, price_type, price, quantity, sold_quantity, created_at
		 FROM listings WHERE status = 'active' ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []model.Listing
	for rows.Next() {
		var l model.Listing
		if err := rows.Scan(&l.ID, &l.ProductID, &l.SellerID, &l.PriceType, &l.Price, &l.Quantity, &l.SoldQty, &l.CreatedAt); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, rows.Err()
}

func (r *ListingRepo) UpdatePrice(id string, newPrice int64) error {
	_, err := r.db.Exec("UPDATE listings SET price = $1 WHERE id = $2", newPrice, id)
	return err
}

func (r *ListingRepo) SoftDelete(id string) error {
	_, err := r.db.Exec("UPDATE listings SET status = 'cancelled' WHERE id = $1", id)
	return err
}
