package repository

import (
	"context"
	"errors"
	"sync"

	"commerce-platform/services/products/internal/product"
	"commerce-platform/shared/logger"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
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

	ErrornousID   = "01a0b072-db8f-742a-a289-0e290e1fb901"
	ErrornousUUID = uuid.MustParse(ErrornousID)
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

func (r *InMemoryProductRepository) FindAll(ctx context.Context) ([]product.Product, error) {
	// dummy way to create unexpected error for tests
	if ctx.Value("errorEnabler") != nil {
		return nil, errors.New(ctx.Value("errorEnabler").(string))
	}

	// read-only: RLock allows concurrent readers.
	r.mu.RLock()
	defer r.mu.RUnlock()

	var products []product.Product

	for _, p := range r.products {
		products = append(products, p)
	}

	return products, nil
}

func (r *InMemoryProductRepository) FindByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	// dummy way to create unexpected error for tests
	if id == ErrornousUUID {
		return product.Product{}, errors.New("unexpected error")
	}

	// read-only: RLock.
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, found := r.products[id]

	if !found {
		return product.Product{}, ErrNotFound
	}

	return p, nil
}

func (r *InMemoryProductRepository) Save(ctx context.Context, p product.Product) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[p.ID] = p
	logger := log(ctx)
	logger.Info().Str("product_id", p.ID.String()).Str("category", p.Category).Msg("product saved")
}

func (r *InMemoryProductRepository) Update(ctx context.Context, p product.Product) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[p.ID] = p
	logger := log(ctx)
	logger.Info().Str("product_id", p.ID.String()).Str("category", p.Category).Msg("product updated")
}

func (r *InMemoryProductRepository) Delete(ctx context.Context, id uuid.UUID) {
	// mutates the map: exclusive Lock.
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.products, id)
	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("product deleted")
}

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "products.repository")
}
