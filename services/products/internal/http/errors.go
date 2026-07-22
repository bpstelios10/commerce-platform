package httpx

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

	switch err {

	case service.ErrProductNotFound:
		writeError(
			w,
			http.StatusNotFound,
			"PRODUCT_NOT_FOUND",
			err.Error(),
		)

	case service.ErrInvalidProduct:
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_PRODUCT",
			err.Error(),
		)

	case service.ErrInvalidCategory:
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_CATEGORY",
			err.Error(),
		)

	case validation.ErrInvalidUUID:
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_UUID",
			err.Error(),
		)

	default:
		logger.Warn().Err(err).Msg("unexpected error handled, with")
		writeError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
	}
}

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
