package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrInvalidOrder        = errors.New("invalid order")
	ErrProductNotFound     = errors.New("product not found for the given id")
	ErrOrderUpdateConflict = errors.New("order was updated by another process")
)

type OrderNotFoundError struct {
	OrderID uuid.UUID
}

func (err *OrderNotFoundError) Error() string {
	return fmt.Sprintf("order with id [%s] was not found", err.OrderID.String())
}
