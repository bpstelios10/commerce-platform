package http

import (
	"commerce-platform/services/orders/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
)

func HandleError(w http.ResponseWriter, err error) {
	// var validationErr ValidationError

	// if errors.As(err, &validationErr) {
	// 	writeError(
	// 		w,
	// 		http.StatusBadRequest,
	// 		"VALIDATION_ERROR",
	// 		validationErr.Error(),
	// 	)
	// 	return
	// }

	switch err {

	case service.ErrOrderNotFound:
		writeError(
			w,
			http.StatusNotFound,
			"ORDER_NOT_FOUND",
			err.Error(),
		)

	case service.ErrInvalidOrder:
		writeError(
			w,
			http.StatusBadRequest,
			"INVALID_ORDER",
			err.Error(),
		)

	default:
		slog.Warn("unexpected error handled, with", "error", err)
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
