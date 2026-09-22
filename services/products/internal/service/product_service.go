package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/shared/logger"
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ProductRepository interface {
	FindAll(ctx context.Context) []product.Product
	FindByID(ctx context.Context, id uuid.UUID) (product.Product, bool)
}

type ProductService struct {
	repository ProductRepository
}

func NewProductService(repository ProductRepository) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) GetProducts(ctx context.Context) []product.Product {
	return s.repository.FindAll(ctx)
}

func (s *ProductService) SearchProducts(ctx context.Context, query string, maxPrice *float64, category string) []product.Product {
	products := s.repository.FindAll(ctx)
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

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	p, found := s.repository.FindByID(ctx, id)
	if !found {
		return product.Product{}, &ErrProductNotFound{ProductID: id}
	}
	return p, nil
}

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "products.service")
}
