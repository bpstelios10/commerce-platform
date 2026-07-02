package repository

import (
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

func (r *InMemoryProductCategoryRepository) Exists(category string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	normalized := strings.ToUpper(strings.TrimSpace(category))
	_, found := r.categories[normalized]
	return found
}
