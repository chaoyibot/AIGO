package product

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/aigo/internal/model"
	"github.com/aigo/internal/repository"
)

type Service struct {
	productRepo *repository.ProductRepo
}

func NewService(productRepo *repository.ProductRepo) *Service {
	return &Service{productRepo: productRepo}
}

func (s *Service) Create(sellerID string, encryptedTitle, encryptedDesc, encryptedMetadata, encryptedKey []byte, priceMin, priceMax int64, category string, tags []string) (*model.Product, error) {
	p := &model.Product{
		ID:                uuid.New().String(),
		SellerID:          sellerID,
		EncryptedTitle:    encryptedTitle,
		EncryptedDesc:     encryptedDesc,
		EncryptedMetadata: encryptedMetadata,
		EncryptedKeySeller: encryptedKey,
		PriceMin:          priceMin,
		PriceMax:          priceMax,
		Category:          category,
		Tags:              tags,
		Status:            "draft",
	}
	if err := s.productRepo.Create(p); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	return p, nil
}

func (s *Service) GetByID(id string) (*model.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *Service) List(filter repository.ProductFilter) ([]model.Product, string, error) {
	return s.productRepo.List(filter)
}

func (s *Service) Update(p *model.Product) error {
	return s.productRepo.Update(p)
}

func (s *Service) Delete(id, sellerID string) error {
	return s.productRepo.SoftDelete(id, sellerID)
}
