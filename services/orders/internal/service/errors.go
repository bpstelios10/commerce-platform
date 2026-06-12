package service

import "errors"

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrInvalidOrder    = errors.New("invalid order")
	ErrProductNotFound = errors.New("product not found for the given id")
)
