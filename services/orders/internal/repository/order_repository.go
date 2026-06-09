package repository

import "commerce-platform/services/orders/internal/order"

type InMemoryOrderRepository struct {
	orders map[string]order.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{
		orders: map[string]order.Order{
			"1": {
				ID:        "1",
				ProductID: "1",
				Quantity:  2,
				Status:    order.CREATED,
			},
			"2": {
				ID:        "2",
				ProductID: "2",
				Quantity:  1,
				Status:    order.PAID,
			},
		},
	}
}

func (repo *InMemoryOrderRepository) FindAll() []order.Order {
	var orders []order.Order

	for _, o := range repo.orders {
		orders = append(orders, o)
	}

	return orders
}

func (repo *InMemoryOrderRepository) FindByID(id string) (order.Order, bool) {
	o, found := repo.orders[id]

	return o, found
}

func (repo *InMemoryOrderRepository) Save(o order.Order) {
	repo.orders[o.ID] = o
}

func (repo *InMemoryOrderRepository) Update(o order.Order) {
	repo.orders[o.ID] = o
}
