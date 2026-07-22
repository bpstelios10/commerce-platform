package logger

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func RequestContextMiddleware(baseLogger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-Id")
			if reqID == "" {
				reqID = uuid.NewString()
				baseLogger.Debug().Str("request_id", reqID).Msg("generated new request id")
			} else {
				baseLogger.Debug().Str("request_id", reqID).Msg("used existing request id")
			}

			// Create a logger tied exclusively to this specific request
			reqLogger := baseLogger.With().
				Str("request_id", reqID).
				Logger()
			ctx := reqLogger.WithContext(r.Context())

			w.Header().Set("X-Request-Id", reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
