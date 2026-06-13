package repository

import (
	"sync"

	"commerce-platform/services/orders/internal/order"
)

var (
	FirstProductID  = "f47ac10b-58cc-4372-a567-0e02b2c3d001"
	SecondProductID = "f47ac10b-58cc-4372-a567-0e02b2c3d002"
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
	orders map[string]order.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: map[string]order.Order{
			"1": {
				ID:        "1",
				ProductID: FirstProductID,
				Quantity:  2,
				Status:    order.CREATED,
			},
			"2": {
				ID:        "2",
				ProductID: SecondProductID,
				Quantity:  1,
				Status:    order.PAID,
			},
		},
	}
}

func (repo *InMemoryOrderRepository) FindAll() []order.Order {
	// read-only: RLock allows concurrent readers.
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	var orders []order.Order

	for _, o := range repo.orders {
		orders = append(orders, o)
	}

	return orders
}

func (repo *InMemoryOrderRepository) FindByID(id string) (order.Order, bool) {
	// read-only: RLock.
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	o, found := repo.orders[id]

	return o, found
}

func (repo *InMemoryOrderRepository) Save(o order.Order) {
	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.orders[o.ID] = o
}

func (repo *InMemoryOrderRepository) Update(o order.Order) {
	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	repo.orders[o.ID] = o
}

func (repo *InMemoryOrderRepository) Delete(id string) {
	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	delete(repo.orders, id)
}
