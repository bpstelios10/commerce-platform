package http

import (
	"context"
	"math"
	"strings"
)

type ValidationError struct {
	Errors []string
}

func (e ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

func validateCreateProduct(ctx context.Context, req CreateProductRequest) error {
	logger := log(ctx)
	validationError := ValidationError{}

	if name := len(strings.TrimSpace(req.Name)); name == 0 || name > 100 {
		validationError.Errors = append(validationError.Errors, "name cannot be blank or exceed 100 characters.")
	}
	if category := len(strings.TrimSpace(req.Category)); category == 0 || category > 50 {
		validationError.Errors = append(validationError.Errors, "category cannot be blank or exceed 50 characters.")
	}
	if !isPriceValid(req.Price) {
		validationError.Errors = append(validationError.Errors, "price must be > 0 and have at most two decimal places.")
	}

	if len(validationError.Errors) > 0 {
		logger.Debug().Strs("errors", validationError.Errors).Msg("invalid create product request")

		return validationError
	}

	return nil
}

func validateUpdateProduct(ctx context.Context, req UpdateProductRequest) error {
	logger := log(ctx)
	validationError := ValidationError{}

	if name := len(strings.TrimSpace(req.Name)); name == 0 || name > 100 {
		validationError.Errors = append(validationError.Errors, "name cannot be blank or exceed 100 characters.")
	}
	if category := len(strings.TrimSpace(req.Category)); category == 0 || category > 50 {
		validationError.Errors = append(validationError.Errors, "category cannot be blank or exceed 50 characters.")
	}
	if !isPriceValid(req.Price) {
		validationError.Errors = append(validationError.Errors, "price must be > 0 and have at most two decimal places.")
	}
	if req.Version <= 0 {
		validationError.Errors = append(validationError.Errors, "version must be > 0.")
	}

	if len(validationError.Errors) > 0 {
		logger.Debug().Strs("errors", validationError.Errors).Msg("invalid update product request")

		return validationError
	}

	return nil
}

func isPriceValid(p float64) bool {
	const maxPrice = 99_999_999.99
	if p <= 0 || p > maxPrice {
		return false
	}
	return math.Round(p*100)/100 == p
}
