package repository

import (
	"context"
	"errors"
	"sync"
	"time"

	"commerce-platform/services/orders/internal/order"
	"commerce-platform/shared/logger"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// InMemoryOrderRepository is shared across goroutines (one instance, called from
// every HTTP request goroutine). Go maps are NOT safe for concurrent use: a write
// happening at the same time as any other access (read or write) panics the process
// with "fatal error: concurrent map read and map write". The mutex serialises access.
//
// We use RWMutex (not plain Mutex) so multiple reads can run in parallel; only writes
// are exclusive. Rule of thumb: reads take RLock, writes take Lock.
type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]order.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: map[uuid.UUID]order.Order{},
	}
}

func (repo *InMemoryOrderRepository) FindAll(ctx context.Context) ([]order.Order, error) {
	// dummy way to create unexpected error for tests
	if ctx.Value("errorEnabler") != nil {
		return nil, errors.New(ctx.Value("errorEnabler").(string))
	}

	// read-only: RLock allows concurrent readers.
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	var orders []order.Order

	for _, o := range repo.orders {
		orders = append(orders, o)
	}

	return orders, nil
}

func (repo *InMemoryOrderRepository) FindByID(ctx context.Context, id uuid.UUID) (order.Order, error) {
	// read-only: RLock.
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	o, found := repo.orders[id]

	if !found {
		return order.Order{}, ErrNotFound
	}

	return o, nil
}

func (repo *InMemoryOrderRepository) Save(ctx context.Context, o order.Order) (order.Order, error) {
	// dummy way to create unexpected error for tests
	if ctx.Value("errorEnabler") != nil {
		return order.Order{}, errors.New(ctx.Value("errorEnabler").(string))
	}

	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	o.CreatedAt = time.Now()
	repo.orders[o.ID] = o
	logger := log(ctx)
	logger.Info().Str("order_id", o.ID.String()).Str("product_id", o.ProductID).Msg("order saved")

	return o, nil
}

func (repo *InMemoryOrderRepository) Update(ctx context.Context, o order.Order) (order.Order, error) {
	// dummy way to create unexpected error for tests
	if ctx.Value("errorEnabler") != nil {
		return order.Order{}, errors.New(ctx.Value("errorEnabler").(string))
	}

	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.orders[o.ID] = o
	logger := log(ctx)
	logger.Info().Str("order_id", o.ID.String()).Str("product_id", o.ProductID).Msg("order updated")
	return o, nil
}

func (repo *InMemoryOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// dummy way to create unexpected error for tests
	if ctx.Value("errorEnabler") != nil {
		return errors.New(ctx.Value("errorEnabler").(string))
	}

	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	delete(repo.orders, id)
	logger := log(ctx)
	logger.Info().Str("order_id", id.String()).Msg("order deleted")

	return nil
}

func log(ctx context.Context) zerolog.Logger {
	return logger.GetLogger(ctx, "orders.repository")
}
