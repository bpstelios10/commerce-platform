package http

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
		msg := fmt.Sprintf("%v", validationError.Errors)

		slog.Warn("invalid create order request", "error", msg)

		return validationError
	}

	return nil
}
