package product

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
	"github.com/aigo/internal/service/crypto"
)

type Service struct {
	productRepo *repository.ProductRepo
	cryptoSvc   *crypto.Service
}

func NewService(productRepo *repository.ProductRepo, cryptoSvc *crypto.Service) *Service {
	return &Service{productRepo: productRepo, cryptoSvc: cryptoSvc}
}

// Create stores the encrypted blobs as-is. Caller is responsible for encrypting
// the title/description/metadata before calling this method.
func (s *Service) Create(sellerID string, encryptedTitle, encryptedDesc, encryptedMetadata, encryptedKey []byte, priceMin, priceMax int64, category string, tags []string) (*model.Product, error) {
	p := &model.Product{
		ID:                 uuid.New().String(),
		SellerID:           sellerID,
		EncryptedTitle:     encryptedTitle,
		EncryptedDesc:      encryptedDesc,
		EncryptedMetadata:  encryptedMetadata,
		EncryptedKeySeller: encryptedKey,
		PriceMin:           priceMin,
		PriceMax:           priceMax,
		Category:           category,
		Tags:               tags,
		Status:             "draft",
	}
	if err := s.productRepo.Create(p); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	return p, nil
}

// GetByID retrieves a product and transparently decrypts the encrypted fields
// so the caller (and ultimately the API consumer) can read the content.
func (s *Service) GetByID(id string) (*model.Product, error) {
	p, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.decryptFields(p); err != nil {
		return nil, fmt.Errorf("decrypt product fields: %w", err)
	}
	return p, nil
}

func (s *Service) List(filter repository.ProductFilter) ([]model.Product, string, error) {
	products, cursor, err := s.productRepo.List(filter)
	if err != nil {
		return nil, "", err
	}
	// Decrypt all products in the list so API consumers can see titles.
	for i := range products {
		_ = s.decryptFields(&products[i]) // ignore decryption errors in list view
	}
	return products, cursor, nil
}

func (s *Service) Update(p *model.Product) error {
	return s.productRepo.Update(p)
}

func (s *Service) Delete(id, sellerID string) error {
	return s.productRepo.SoftDelete(id, sellerID)
}

// decryptFields decrypts the encrypted blobs in-place and populates the
// Decrypted* fields. Fields that are empty byte slices are left as empty strings.
func (s *Service) decryptFields(p *model.Product) error {
	if len(p.EncryptedTitle) > 0 {
		plain, err := s.cryptoSvc.Decrypt(p.EncryptedTitle)
		if err != nil {
			return fmt.Errorf("decrypt title: %w", err)
		}
		p.DecryptedTitle = string(plain)
	}
	if len(p.EncryptedDesc) > 0 {
		plain, err := s.cryptoSvc.Decrypt(p.EncryptedDesc)
		if err != nil {
			return fmt.Errorf("decrypt description: %w", err)
		}
		p.DecryptedDesc = string(plain)
	}
	if len(p.EncryptedMetadata) > 0 {
		plain, err := s.cryptoSvc.Decrypt(p.EncryptedMetadata)
		if err != nil {
			return fmt.Errorf("decrypt metadata: %w", err)
		}
		p.DecryptedMetadata = string(plain)
	}
	return nil
}
