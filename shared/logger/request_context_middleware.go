package logger

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// statusRecordingResponseWriter captures the status code written by the handler so it
// can be included in the canonical "request completed" log line below.
type statusRecordingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusRecordingResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func RequestContextMiddleware(baseLogger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqID := r.Header.Get("X-Request-Id")
			if reqID == "" {
				reqID = uuid.NewString()
			}

			// Create a logger tied exclusively to this specific request
			reqLogger := baseLogger.With().
				Str("request_id", reqID).
				Logger()
			ctx := reqLogger.WithContext(r.Context())
			ctx = ContextWithRequestID(ctx, reqID)

			w.Header().Set("X-Request-Id", reqID)

			recorder := &statusRecordingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r.WithContext(ctx))

			// Canonical, single per-request access log line - independent of whatever
			// business-event logs individual handlers add.
			reqLogger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", recorder.status).
				Dur("duration", time.Since(start)).
				Msg("request completed")
		})
	}
}
