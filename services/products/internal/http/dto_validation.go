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

	if len(strings.TrimSpace(req.Name)) == 0 {
		validationError.Errors = append(validationError.Errors, "name cannot be blank.")
	}
	if len(strings.TrimSpace(req.Category)) == 0 {
		validationError.Errors = append(validationError.Errors, "category cannot be blank.")
	}
	if req.Price <= 0 {
		validationError.Errors = append(validationError.Errors, "price must be > 0.")
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

	if len(strings.TrimSpace(req.Name)) == 0 {
		validationError.Errors = append(validationError.Errors, "name cannot be blank.")
	}
	if len(strings.TrimSpace(req.Category)) == 0 {
		validationError.Errors = append(validationError.Errors, "category cannot be blank.")
	}
	if req.Price <= 0 {
		validationError.Errors = append(validationError.Errors, "price must be > 0.")
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
	if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > maxPrice {
		return false
	}
	return math.Round(p*100)/100 == p
}
