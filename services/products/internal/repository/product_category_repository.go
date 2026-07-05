package repository

import (
	"strings"
	"sync"

	"log/slog"
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

func (r *InMemoryProductCategoryRepository) Exists(category string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ToUpper(strings.TrimSpace(category))
	_, found := r.categories[normalized]
	slog.Info("does product category exist?", "category", category, "exists", found)

	return found
}

func (r *InMemoryProductCategoryRepository) GetAll() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	categoriesNames := make([]string, 0, len(r.categories))
	for category := range r.categories {
		categoriesNames = append(categoriesNames, category)
	}
	slog.Info("product categories retrieved", "categories", categoriesNames)

	return categoriesNames
}
