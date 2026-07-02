package service

import (
	"commerce-platform/services/products/internal/product"
	"log/slog"
	"strings"

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

func (s *ProductService) SearchProducts(query string, maxPrice *float64, category string) []product.Product {
	products := s.repository.FindAll()
	query = strings.ToLower(strings.TrimSpace(query))
	category = strings.ToLower(strings.TrimSpace(category))

	filtered := make([]product.Product, 0, len(products))

	for _, p := range products {
		if query != "" && !strings.Contains(strings.ToLower(p.Name), query) {
			continue
		}

		if maxPrice != nil && p.Price > *maxPrice {
			continue
		}

		if category != "" && strings.ToLower(p.Category) != category {
			continue
		}

		filtered = append(filtered, p)
	}

	return filtered
}

func (s *ProductService) GetProductByID(id uuid.UUID) (product.Product, error) {
	p, found := s.repository.FindByID(id)
	if !found {
		slog.Warn("product error for", "productId", id, "error", ErrProductNotFound)
		return product.Product{}, ErrProductNotFound
	}
	return p, nil
}
