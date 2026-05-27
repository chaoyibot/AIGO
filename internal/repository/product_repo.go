package repository

import (
	"database/sql"
	"fmt"
	"github.com/aigo/internal/model"
)

type ProductRepo struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(p *model.Product) error {
	_, err := r.db.Exec(
		`INSERT INTO products (id, seller_id, encrypted_title, encrypted_description, encrypted_metadata, encrypted_key_seller, price_min, price_max, category, tags, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		p.ID, p.SellerID, p.EncryptedTitle, p.EncryptedDesc, p.EncryptedMetadata, p.EncryptedKeySeller,
		p.PriceMin, p.PriceMax, p.Category, p.Tags, p.Status)
	return err
}

func (r *ProductRepo) FindByID(id string) (*model.Product, error) {
	p := &model.Product{}
	err := r.db.QueryRow(
		`SELECT id, seller_id, encrypted_title, encrypted_description, encrypted_metadata, encrypted_key_seller, price_min, price_max, category, tags, status, created_at, updated_at
		 FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.SellerID, &p.EncryptedTitle, &p.EncryptedDesc, &p.EncryptedMetadata, &p.EncryptedKeySeller,
		&p.PriceMin, &p.PriceMax, &p.Category, &p.Tags, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("find product: %w", err)
	}
	return p, nil
}

type ProductFilter struct {
	Category string
	Status   string
	PriceMin int64
	PriceMax int64
	Tags     []string
	Limit    int
	Cursor   string // base64 encoded last id
}

func (r *ProductRepo) List(filter ProductFilter) ([]model.Product, string, error) {
	query := `SELECT id, seller_id, price_min, price_max, category, tags, status, created_at
		 FROM products WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if filter.Category != "" {
		query += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, filter.Category)
		argIdx++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	} else {
		query += " AND status = 'active'"
	}
	if filter.PriceMin > 0 {
		query += fmt.Sprintf(" AND price_max >= $%d", argIdx)
		args = append(args, filter.PriceMin)
		argIdx++
	}
	if filter.PriceMax > 0 {
		query += fmt.Sprintf(" AND price_min <= $%d", argIdx)
		args = append(args, filter.PriceMax)
		argIdx++
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit+1) // +1 to check has_more

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.SellerID, &p.PriceMin, &p.PriceMax, &p.Category, &p.Tags, &p.Status, &p.CreatedAt); err != nil {
			return nil, "", err
		}
		products = append(products, p)
	}

	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	nextCursor := ""
	if hasMore && len(products) > 0 {
		nextCursor = products[len(products)-1].ID
	}

	return products, nextCursor, nil
}

func (r *ProductRepo) Update(p *model.Product) error {
	_, err := r.db.Exec(
		`UPDATE products SET encrypted_title=$1, encrypted_description=$2, encrypted_metadata=$3, price_min=$4, price_max=$5, category=$6, tags=$7, status=$8, updated_at=now()
		 WHERE id=$9 AND seller_id=$10`,
		p.EncryptedTitle, p.EncryptedDesc, p.EncryptedMetadata, p.PriceMin, p.PriceMax, p.Category, p.Tags, p.Status, p.ID, p.SellerID)
	return err
}

func (r *ProductRepo) SoftDelete(id, sellerID string) error {
	_, err := r.db.Exec("UPDATE products SET status='archived', updated_at=now() WHERE id=$1 AND seller_id=$2", id, sellerID)
	return err
}
