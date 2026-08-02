package order

import (
	"strings"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	CREATED  OrderStatus = "CREATED"
	PAID     OrderStatus = "PAID"
	RETURNED OrderStatus = "RETURNED"
	CANCELED OrderStatus = "CANCELED"
)

type Order struct {
	ID        uuid.UUID   `json:"id"`
	ProductID string      `json:"product_id"`
	Quantity  int         `json:"quantity"`
	Status    OrderStatus `json:"status"`
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case CREATED, PAID, RETURNED, CANCELED:
		return true
	}
	return false
}

func (s OrderStatus) Normalize() OrderStatus {
	return OrderStatus(strings.ToUpper(string(s)))
}
