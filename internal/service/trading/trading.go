package trading

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type Service struct {
	db          *sql.DB
	productRepo *repository.ProductRepo
	listingRepo *repository.ListingRepo
	walletRepo  *repository.WalletRepo
}

func NewService(db *sql.DB, productRepo *repository.ProductRepo, listingRepo *repository.ListingRepo, walletRepo *repository.WalletRepo) *Service {
	return &Service{db: db, productRepo: productRepo, listingRepo: listingRepo, walletRepo: walletRepo}
}

// BuyListing handles the full purchase transaction:
// 1. Lock listing row (FOR UPDATE)
// 2. Check stock
// 3. Lock buyer wallet
// 4. Deduct points
// 5. Update sold quantity
// 6. Create order
// 7. Record transaction
func (s *Service) BuyListing(buyerID, listingID string, quantity int) (*model.Order, error) {
	if quantity <= 0 {
		quantity = 1
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Lock listing row
	listing := &model.Listing{}
	err = tx.QueryRow(
		`SELECT id, product_id, seller_id, price_type, price, quantity, sold_quantity, status, expires_at
		 FROM listings WHERE id = $1 FOR UPDATE`,
		listingID,
	).Scan(&listing.ID, &listing.ProductID, &listing.SellerID, &listing.PriceType, &listing.Price, &listing.Quantity, &listing.SoldQty, &listing.Status, &listing.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("listing not found")
		}
		return nil, fmt.Errorf("lock listing: %w", err)
	}

	if listing.Status != "active" {
		return nil, fmt.Errorf("listing not active")
	}
	if listing.SellerID == buyerID {
		return nil, fmt.Errorf("cannot buy your own listing")
	}

	remaining := listing.Quantity - listing.SoldQty
	if remaining < quantity {
		return nil, fmt.Errorf("insufficient stock: %d remaining, %d requested", remaining, quantity)
	}

	// 2. Lock buyer wallet
	buyerWallet, err := s.walletRepo.FindByUserIDForUpdate(tx, buyerID)
	if err != nil {
		return nil, fmt.Errorf("buyer wallet not found")
	}

	totalPrice := listing.Price * int64(quantity)
	if buyerWallet.PointsBalance < totalPrice {
		return nil, fmt.Errorf("insufficient points: have %d, need %d", buyerWallet.PointsBalance, totalPrice)
	}

	// 3. Deduct points from buyer
	_, err = tx.Exec(
		"UPDATE wallets SET points_balance = points_balance - $1, updated_at = now() WHERE user_id = $2",
		totalPrice, buyerID,
	)
	if err != nil {
		return nil, fmt.Errorf("deduct buyer points: %w", err)
	}

	// 4. Add points to seller (create wallet if not exists via upsert)
	_, err = tx.Exec(
		`INSERT INTO wallets (id, user_id, points_balance, currency)
		 VALUES ($1, $2, $3, 'CNY')
		 ON CONFLICT (user_id)
		 DO UPDATE SET points_balance = wallets.points_balance + $3, updated_at = now()`,
		uuid.New().String(), listing.SellerID, totalPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("credit seller points: %w", err)
	}

	// 5. Update sold quantity on listing
	_, err = tx.Exec("UPDATE listings SET sold_quantity = sold_quantity + $1 WHERE id = $2", quantity, listingID)
	if err != nil {
		return nil, fmt.Errorf("update sold quantity: %w", err)
	}

	// 6. Mark listing sold if fully sold out
	newSold := listing.SoldQty + quantity
	if newSold >= listing.Quantity {
		_, err = tx.Exec("UPDATE listings SET status = 'sold' WHERE id = $1", listingID)
		if err != nil {
			return nil, fmt.Errorf("mark listing sold: %w", err)
		}
	}

	// 7. Create order record
	order := &model.Order{
		ID:           uuid.New().String(),
		ListingID:    listingID,
		BuyerID:      buyerID,
		SellerID:     listing.SellerID,
		Quantity:     quantity,
		UnitPrice:    listing.Price,
		TotalPrice:   totalPrice,
		Status:       "paid",
		EscrowStatus: "released",
	}

	_, err = tx.Exec(
		`INSERT INTO orders (id, listing_id, buyer_id, seller_id, quantity, unit_price, total_price, status, escrow_status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		order.ID, listingID, buyerID, listing.SellerID, quantity, listing.Price, totalPrice, "paid", "released",
	)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// 8. Record transaction for buyer
	_, err = tx.Exec(
		`INSERT INTO transactions (id, user_id, type, amount, balance, reference)
		 VALUES ($1, $2, 'purchase', $3, $4, $5)`,
		uuid.New().String(), buyerID, -totalPrice, buyerWallet.PointsBalance-totalPrice, "购买商品:"+listingID,
	)
	if err != nil {
		return nil, fmt.Errorf("record transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("buy commit: %w", err)
	}

	return order, nil
}
