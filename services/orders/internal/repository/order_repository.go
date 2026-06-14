package repository

import (
	"sync"

	"commerce-platform/services/orders/internal/order"

	"github.com/google/uuid"
)

var (
	FirstProductID  = "f47ac10b-58cc-4372-a567-0e02b2c3d001"
	SecondProductID = "f47ac10b-58cc-4372-a567-0e02b2c3d002"

	FirstOrderID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d011")
	SecondOrderID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d012")
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
		orders: map[uuid.UUID]order.Order{
			FirstOrderID: {
				ID:        FirstOrderID,
				ProductID: FirstProductID,
				Quantity:  2,
				Status:    order.CREATED,
			},
			SecondOrderID: {
				ID:        SecondOrderID,
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

func (repo *InMemoryOrderRepository) FindByID(id uuid.UUID) (order.Order, bool) {
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

func (repo *InMemoryOrderRepository) Delete(id uuid.UUID) {
	// mutates the map: exclusive Lock.
	repo.mu.Lock()
	defer repo.mu.Unlock()

	delete(repo.orders, id)
}
