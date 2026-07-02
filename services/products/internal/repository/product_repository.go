package repository

import (
	"sync"

	"commerce-platform/services/products/internal/product"

	"github.com/google/uuid"
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
	products map[uuid.UUID]product.Product
}

var (
	FirstUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001")
	SecondUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002")
	ThirdUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d003")
	FourthUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d004")
)

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: map[uuid.UUID]product.Product{
			FirstUUID: {
				ID:       FirstUUID,
				Name:     "MacBook Pro",
				Category: "ACCESSORY",
				Price:    2500,
				Stock:    10,
			},
			SecondUUID: {
				ID:       SecondUUID,
				Name:     "iPhone",
				Category: "ACCESSORY",
				Price:    1200,
				Stock:    5,
			},
			ThirdUUID: {
				ID:       ThirdUUID,
				Name:     "hoodie Mykonos",
				Category: "CLOTHES",
				Price:    80,
				Stock:    8,
			},
			FourthUUID: {
				ID:       FourthUUID,
				Name:     "Eye necklace",
				Category: "JEWELRY",
				Price:    150,
				Stock:    15,
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

func (r *InMemoryProductRepository) FindByID(id uuid.UUID) (product.Product, bool) {
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

func (r *InMemoryProductRepository) Delete(id uuid.UUID) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.products, id)
}
