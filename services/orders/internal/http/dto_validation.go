package http

import (
	"commerce-platform/services/orders/internal/validation"
	"context"
	"strings"
)

type ValidationError struct {
	Errors []string
}

func (e ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

func validateCreateOrder(ctx context.Context, req CreateOrderRequest) error {
	logger := log(ctx)
	validationError := ValidationError{}

	trimmedProductId := strings.TrimSpace(req.ProductID)
	if len(trimmedProductId) == 0 {
		validationError.Errors = append(validationError.Errors, "product-id cannot be blank.")
	} else if _, err := validation.GetValidUUID(trimmedProductId); err != nil {
		validationError.Errors = append(validationError.Errors, "product-id is not a valid UUID.")
	}
	if req.Quantity <= 0 {
		validationError.Errors = append(validationError.Errors, "quantity must be > 0.")
	}

	if len(validationError.Errors) > 0 {
		logger.Debug().Strs("errors", validationError.Errors).Msg("invalid create order request")

		return validationError
	}

	return nil
}

func validateUpdateOrder(ctx context.Context, req UpdateOrderRequest) error {
	logger := log(ctx)
	validationError := ValidationError{}

	trimmedProductId := strings.TrimSpace(req.ProductID)
	if len(trimmedProductId) == 0 {
		validationError.Errors = append(validationError.Errors, "product-id cannot be blank.")
	} else if _, err := validation.GetValidUUID(trimmedProductId); err != nil {
		validationError.Errors = append(validationError.Errors, "product-id is not a valid UUID.")
	}
	if req.Quantity <= 0 {
		validationError.Errors = append(validationError.Errors, "quantity must be > 0.")
	}
	if !req.Status.IsValid() {
		validationError.Errors = append(validationError.Errors, "status is not valid.")
	}

	if len(validationError.Errors) > 0 {
		logger.Debug().Strs("errors", validationError.Errors).Msg("invalid update order request")

		return validationError
	}

	return nil
}
