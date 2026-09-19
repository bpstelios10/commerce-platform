package http

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := log(ctx)
	var validationErr ValidationError

	if errors.As(err, &validationErr) {
		writeError(
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
			w,
			http.StatusNotFound,
			"PRODUCT_NOT_FOUND",
			service.ErrProductNotFound.Error(),
		)

	case errors.Is(err, service.ErrInvalidProduct):
		logger.Warn().Err(err).Msg("invalid product")
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_PRODUCT",
			service.ErrInvalidProduct.Error(),
		)

	case errors.Is(err, service.ErrInvalidCategory):
		logger.Warn().Err(err).Msg("invalid category")
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_CATEGORY",
			service.ErrInvalidCategory.Error(),
		)

	case errors.Is(err, validation.ErrInvalidUUID):
		logger.Warn().Err(err).Msg("invalid UUID")
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_UUID",
			validation.ErrInvalidUUID.Error(),
		)

	default:
		logger.Error().Err(err).Msg("unexpected error handled, with")
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
	}
}

// TODO move to utils as handler-error-response?
func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(
		ErrorResponse{
			Code:    code,
			Message: message,
		},
	)
}
