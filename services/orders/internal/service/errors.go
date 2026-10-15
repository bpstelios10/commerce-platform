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

type ErrOrderNotFound struct {
	OrderID uuid.UUID
}

func (err *ErrOrderNotFound) Error() string {
	return fmt.Sprintf("Order with id [%s] was not found", err.OrderID.String())
}
