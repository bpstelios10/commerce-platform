package http

import (
	"log/slog"
	"strings"
)

type ValidationError struct {
	Errors []string
}

func (e ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

func validateCreateOrder(req CreateOrderRequest) error {
	validationError := ValidationError{}

	if len(strings.TrimSpace(req.ID)) == 0 {
		validationError.Errors = append(validationError.Errors, "id cannot be blank.")
	}
	if len(strings.TrimSpace(req.ProductID)) == 0 {
		validationError.Errors = append(validationError.Errors, "product-id cannot be blank.")
	}
	if req.Quantity <= 0 {
		validationError.Errors = append(validationError.Errors, "quantity must be > 0.")
	}

	if len(validationError.Errors) > 0 {
		slog.Warn("invalid create order request", "errors", validationError.Errors)

		return validationError
	}

	return nil
}

func validateUpdateOrder(req UpdateOrderRequest) error {
	validationError := ValidationError{}

	if len(strings.TrimSpace(req.ProductID)) == 0 {
		validationError.Errors = append(validationError.Errors, "product-id cannot be blank.")
	}
	if req.Quantity <= 0 {
		validationError.Errors = append(validationError.Errors, "quantity must be > 0.")
	}
	if !req.Status.IsValid() {
		validationError.Errors = append(validationError.Errors, "status is not valid.")
	}

	if len(validationError.Errors) > 0 {
		slog.Warn("invalid create order request", "errors", validationError.Errors)

		return validationError
	}

	return nil
}
