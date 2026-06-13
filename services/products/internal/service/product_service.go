package service

import (
	"commerce-platform/services/products/internal/product"
	"log/slog"

	"github.com/google/uuid"
)

type ProductRepository interface {
	FindAll() []product.Product
	FindByID(id uuid.UUID) (product.Product, bool)
}

type ProductService struct {
	repository ProductRepository
}

func NewProductService(repository ProductRepository) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) GetProducts() []product.Product {
	return s.repository.FindAll()
}

func (s *ProductService) GetProductByID(id uuid.UUID) (product.Product, error) {
	p, found := s.repository.FindByID(id)
	if !found {
		slog.Warn("product error for", "productId", id, "error", ErrProductNotFound)
		return product.Product{}, ErrProductNotFound
	}
	return p, nil
}
