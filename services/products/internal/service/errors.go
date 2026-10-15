package service

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrInvalidProduct        = errors.New("invalid product")
	ErrInvalidCategory       = errors.New("invalid category")
	ErrProductUpdateConflict = errors.New("product was updated by another process")
)

type ErrProductNotFound struct {
	ProductID uuid.UUID
}

func (err *ErrProductNotFound) Error() string {
	return fmt.Sprintf("Product with id [%s] was not found", err.ProductID.String())
}
