package service

import "errors"

var (
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidProduct  = errors.New("invalid product")
	ErrInvalidCategory = errors.New("invalid category")
)
