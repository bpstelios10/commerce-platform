package order

type OrderStatus string

const (
	CREATED   OrderStatus = "CREATED"
	PAID      OrderStatus = "PAID"
	RETURNED  OrderStatus = "RETURNED"
	CANCELLED OrderStatus = "CANCELLED"
)

type Order struct {
	ID        string      `json:"id"`
	ProductID string      `json:"product_id"`
	Quantity  int         `json:"quantity"`
	Status    OrderStatus `json:"status"`
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case CREATED, PAID, RETURNED, CANCELLED:
		return true
	}
	return false
}
