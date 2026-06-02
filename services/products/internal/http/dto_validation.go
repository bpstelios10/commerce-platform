package httpx

import (
	"fmt"
	"log/slog"
	"strings"
)

type ValidationError struct {
	Errors []string
}

func (e ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

func validateCreateProduct(req CreateProductRequest) error {
	validationError := ValidationError{}

	if len(strings.TrimSpace(req.ID)) == 0 {
		validationError.Errors = append(validationError.Errors, "id cannot be blank.")
	}
	if len(strings.TrimSpace(req.Name)) == 0 {
		validationError.Errors = append(validationError.Errors, "name cannot be blank.")
	}
	if req.Price <= 0 {
		validationError.Errors = append(validationError.Errors, "price must be > 0.")
	}

	if len(validationError.Errors) > 0 {
		msg := fmt.Sprintf("%v", validationError.Errors)

		slog.Warn("invalid create product request", "error", msg)

		return validationError
	}

	return nil
}

func validateUpdateProduct(req UpdateProductRequest) error {
	validationError := ValidationError{}

	if len(strings.TrimSpace(req.Name)) == 0 {
		validationError.Errors = append(validationError.Errors, "name cannot be blank.")
	}
	if req.Price <= 0 {
		validationError.Errors = append(validationError.Errors, "price must be > 0.")
	}

	if len(validationError.Errors) > 0 {
		msg := fmt.Sprintf("%v", validationError.Errors)

		slog.Warn("invalid update product request", "error", msg)

		return validationError
	}

	return nil
}
