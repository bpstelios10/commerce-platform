package repository

import (
	"sync"

	"commerce-platform/services/products/internal/product"
)

// InMemoryProductRepository is shared across goroutines (one instance, called from
// every HTTP request goroutine). Go maps are NOT safe for concurrent use: a write
// happening at the same time as any other access (read or write) panics the process
// with "fatal error: concurrent map read and map write". The mutex serialises access.
//
// We use RWMutex (not plain Mutex) so multiple reads can run in parallel; only writes
// are exclusive. Rule of thumb: reads take RLock, writes take Lock.
type InMemoryProductRepository struct {
	mu       sync.RWMutex
	products map[string]product.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: map[string]product.Product{
			"1": {
				ID:    "1",
				Name:  "MacBook Pro",
				Price: 2500,
			},
			"2": {
				ID:    "2",
				Name:  "iPhone",
				Price: 1200,
			},
		},
	}
}

func (r *InMemoryProductRepository) FindAll() []product.Product {
	// read-only: RLock allows concurrent readers.
	r.mu.RLock()
	defer r.mu.RUnlock()

	var products []product.Product

	for _, p := range r.products {
		products = append(products, p)
	}

	return products
}

func (r *InMemoryProductRepository) FindByID(id string) (product.Product, bool) {
	// read-only: RLock.
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, found := r.products[id]
	return p, found
}

func (r *InMemoryProductRepository) Save(p product.Product) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[p.ID] = p
}

func (r *InMemoryProductRepository) Update(p product.Product) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[p.ID] = p
}

func (r *InMemoryProductRepository) Delete(id string) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.products, id)
}
