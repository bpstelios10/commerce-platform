package http

import (
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/services/orders/internal/validation"
	"context"
	"errors"
	"fmt"
	"net/http"
)

type TooLongBodyErr struct {
	limit int64
}

func (err TooLongBodyErr) Error() string {
	return fmt.Sprintf("Request body exceeds %d bytes", err.limit)
}

func HandleError(ctx context.Context, w http.ResponseWriter, err error) {
	logger := log(ctx)

	var validationErr ValidationError
	var notFoundErr *service.ErrOrderNotFound
	var tooLongBodyErr TooLongBodyErr

	switch {

	case errors.As(err, &validationErr):
		writeError(ctx, w, http.StatusBadRequest, "VALIDATION_ERROR", validationErr.Error())

	case errors.As(err, &notFoundErr):
		logger.Warn().Err(err).Msg("order not found")
		writeError(ctx, w, http.StatusNotFound, "ORDER_NOT_FOUND", notFoundErr.Error())

	case errors.Is(err, service.ErrInvalidOrder):
		logger.Warn().Err(err).Msg("invalid order")
		writeError(ctx, w, http.StatusBadRequest, "INVALID_ORDER", service.ErrInvalidOrder.Error())

	case errors.Is(err, service.ErrProductNotFound):
		logger.Warn().Err(err).Msg("product not found")
		writeError(ctx, w, http.StatusConflict, "PRODUCT_NOT_FOUND", service.ErrProductNotFound.Error())

	case errors.Is(err, validation.ErrInvalidUUID):
		logger.Warn().Err(err).Msg("invalid UUID")
		writeError(ctx, w, http.StatusBadRequest, "INVALID_UUID", err.Error())

	case errors.As(err, &tooLongBodyErr):
		logger.Warn().Err(err).Msg("Request body too long")
		writeError(ctx, w, http.StatusBadRequest, "REQUEST_BODY_TOO_LONG", tooLongBodyErr.Error())

	default:
		logger.Error().Err(err).Msg("unexpected error handled, with")
		writeError(ctx, w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "internal server error")
	}
}

func writeError(ctx context.Context, w http.ResponseWriter, status int, code string, message string) {
	body := ErrorResponse{
		Code:    code,
		Message: message,
	}

	HandleResponseWithBody(ctx, w, status, body)
}
