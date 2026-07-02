package repository

import (
	"sync"

	"commerce-platform/services/products/internal/product"
)

type InMemoryProductCategoryRepository struct {
	mu         sync.RWMutex
	categories map[product.ProductCategory]struct{}
}

func NewInMemoryProductCategoryRepository() *InMemoryProductCategoryRepository {
	return &InMemoryProductCategoryRepository{
		categories: map[product.ProductCategory]struct{}{
			product.ProductCategory("MAGNET"):    {},
			product.ProductCategory("POSTCARD"):  {},
			product.ProductCategory("ACCESSORY"): {},
			product.ProductCategory("JEWELRY"):   {},
			product.ProductCategory("CLOTHES"):   {},
		},
	}
}

func (r *InMemoryProductCategoryRepository) Exists(category product.ProductCategory) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, found := r.categories[category]
	return found
}
