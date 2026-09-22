package http

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"context"
	"errors"
	"net/http"
)

func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := log(ctx)
	var validationErr ValidationError

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

	case errors.Is(err, service.ErrProductNotFound):
		logger.Warn().Err(err).Msg("product not found")
		writeError(
			ctx,
			w,
			http.StatusNotFound,
			"PRODUCT_NOT_FOUND",
			service.ErrProductNotFound.Error(),
		)

	case errors.Is(err, service.ErrInvalidProduct):
		logger.Warn().Err(err).Msg("invalid product")
		writeError(
			ctx,
			w,
			http.StatusBadRequest,
			"INVALID_PRODUCT",
			service.ErrInvalidProduct.Error(),
		)

	case errors.Is(err, service.ErrInvalidCategory):
		logger.Warn().Err(err).Msg("invalid product category")
		writeError(
			ctx,
			w,
			http.StatusBadRequest,
			"INVALID_CATEGORY",
			service.ErrInvalidCategory.Error(),
		)

	case errors.Is(err, validation.ErrInvalidUUID):
		logger.Warn().Err(err).Msg("invalid UUID")
		writeError(
			ctx,
			w,
			http.StatusBadRequest,
			"INVALID_UUID",
			validation.ErrInvalidUUID.Error(),
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
