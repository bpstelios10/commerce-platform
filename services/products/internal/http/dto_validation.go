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

	if req.ID == "" {
		validationError.Errors = append(validationError.Errors, "id is required.")
	}
	if req.Name == "" {
		validationError.Errors = append(validationError.Errors, "name is required.")
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

	if req.Name == "" {
		validationError.Errors = append(validationError.Errors, "name is required.")
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
