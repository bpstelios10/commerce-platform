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

type ProductNotFoundError struct {
	ProductID uuid.UUID
}

func (err *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product with id [%s] was not found", err.ProductID.String())
}
