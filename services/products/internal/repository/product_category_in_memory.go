package repository

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type InMemoryProductCategoryRepository struct {
	mu         sync.RWMutex
	categories map[string]struct{}
}

func NewInMemoryProductCategoryRepository() *InMemoryProductCategoryRepository {
	return &InMemoryProductCategoryRepository{
		categories: map[string]struct{}{
			"MAGNET":    {},
			"POSTCARD":  {},
			"ACCESSORY": {},
			"JEWELRY":   {},
			"CLOTHES":   {},
		},
	}
}

func (r *InMemoryProductCategoryRepository) Exists(ctx context.Context, category string) (bool, error) {
	// dummy way to create unexpected error for tests
	if category == "ERRORNOUS_CATEGORY" {
		return false, errors.New("unexpected error")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ToUpper(strings.TrimSpace(category))
	_, found := r.categories[normalized]
	logger := log(ctx)
	logger.Info().Str("category", normalized).Bool("exists", found).Msg("checked product category")

	return found, nil
}

func (r *InMemoryProductCategoryRepository) GetAll(ctx context.Context) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoriesNames := make([]string, 0, len(r.categories))
	for category := range r.categories {
		categoriesNames = append(categoriesNames, category)
	}
	logger := log(ctx)
	logger.Info().Strs("categories", categoriesNames).Msg("product categories retrieved")

	return categoriesNames
}
