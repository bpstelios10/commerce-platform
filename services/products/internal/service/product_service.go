package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/shared/logger"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ProductRepository interface {
	List(ctx context.Context) ([]product.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (product.Product, error)
}

type ProductServiceImpl struct {
	repository ProductRepository
}

func NewProductServiceImpl(repository ProductRepository) *ProductServiceImpl {
	return &ProductServiceImpl{
		repository: repository,
	}
}

func (s *ProductServiceImpl) GetProducts(ctx context.Context) ([]product.Product, error) {
	products, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

	return products, nil
}

// TODO implement it using postgre features such as:
// pg_trgm (trigram) → fuzzy/partial/typo tolerance (widgit ≈ widget).
// Full-text search (tsvector) → handles plural/stemming ("widgets" → "widget") automatically.
// GIN (Generalized Inverted Index) → efficient indexing for multiple columns, useful for combined search scenarios.
func (s *ProductServiceImpl) SearchProducts(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
	products, err := s.repository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("get products: %w", err)
	}

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

	return filtered, nil
}

func (s *ProductServiceImpl) GetProductByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	p, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return product.Product{}, fmt.Errorf("get product by id: %w", &ErrProductNotFound{ProductID: id})
		}

		return product.Product{}, fmt.Errorf("get product by id: %w", err)
	}

	return p, nil
}

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "products.service")
}
