package http

import (
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/internal/validation"
	"context"
	"errors"
	"net/http"
)

func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	var validationErr ValidationError
	logger := log(ctx)

	if errors.As(err, &validationErr) {
		writeError(
			ctx,
			w,
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			validationErr.Error(),
		)
		return
	}

	switch {

	case errors.Is(err, service.ErrOrderNotFound):
		logger.Warn().Err(err).Msg("order not found")
		writeError(
			ctx,
			w,
			http.StatusNotFound,
			"ORDER_NOT_FOUND",
			service.ErrOrderNotFound.Error(),
		)

	case errors.Is(err, service.ErrInvalidOrder):
		logger.Warn().Err(err).Msg("invalid order")
		writeError(
			ctx,
			w,
			http.StatusBadRequest,
			"INVALID_ORDER",
			service.ErrInvalidOrder.Error(),
		)

	case errors.Is(err, service.ErrProductNotFound):
		logger.Warn().Err(err).Msg("product not found")
		writeError(
			ctx,
			w,
			http.StatusConflict,
			"PRODUCT_NOT_FOUND",
			service.ErrProductNotFound.Error(),
		)

	case errors.Is(err, validation.ErrInvalidUUID):
		logger.Warn().Err(err).Msg("invalid UUID")
		writeError(
			ctx,
			w,
			http.StatusBadRequest,
			"INVALID_UUID",
			err.Error(),
		)

	default:
		logger.Error().Err(err).Msg("unexpected error handled, with")
		writeError(
			ctx,
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
	}
}

func writeError(ctx context.Context, w http.ResponseWriter, status int, code string, message string) {
	body := ErrorResponse{
		Code:    code,
		Message: message,
	}

	HandleResponseWithBody(ctx, w, status, body)
}
