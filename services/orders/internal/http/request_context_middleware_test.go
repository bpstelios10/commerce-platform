package http

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/repository"
	"commerce-platform/services/orders/internal/service"
	"commerce-platform/shared/logger"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type mockProductsClient2 struct {
	productIDs map[string]bool
}

func (m *mockProductsClient2) GetProductByID(_ context.Context, id string) (*grpc.GetProductByIDResponse, error) {
	if m.productIDs[id] {
		return &grpc.GetProductByIDResponse{Id: id}, nil
	}
	return nil, service.ErrProductNotFound
}

// To be used as BeforeEach
func setupOrderHandlerWithContextTest(t *testing.T) (*chi.Mux, *repository.InMemoryOrderRepository) {
	t.Helper()
	repo := repository.NewInMemoryOrderRepository()
	client := &mockProductsClient2{
		productIDs: map[string]bool{
			repository.FirstProductID:  true,
			repository.SecondProductID: true,
		},
	}
	svc := service.NewOrderService(repo, client)
	handler := NewOrderHandler(svc)

	logger := logger.New(logger.Config{
		Service: "orders",
		Env:     "local",
		Level:   zerolog.InfoLevel,
	})
	// set the default slog to point to logger, just in case
	slogHandler := zerolog.NewSlogHandler(logger)
	slog.SetDefault(slog.New(slogHandler))

	logger.Info().Msg("Commerce Platform - ORDERS")
	r := chi.NewRouter()
	r.Use(RequestContextMiddleware(logger))
	handler.RegisterRoutes(r)

	return r, repo
}

func TestRequestContextMiddleware_WhenRequestIdProvided_ReturnsSameInResponseHeaders(t *testing.T) {
	r, _ := setupOrderHandlerWithContextTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/"+repository.FirstOrderID.String(),
		nil,
	)
	req.Header.Set("X-Request-Id", "test-request-id")
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.Equal(t, "test-request-id", res.Header().Get("X-Request-Id"))
	assert.JSONEq(
		t,
		`{
			"id":         "`+repository.FirstOrderID.String()+`",
			"product_id": "`+repository.FirstProductID+`",
			"quantity":   2,
			"status":     "CREATED"
		}`,
		res.Body.String(),
	)
}

func TestRequestContextMiddleware_WhenRequestIdNotProvided_ReturnsNewInResponseHeadersAndIsUuid(t *testing.T) {
	r, _ := setupOrderHandlerWithContextTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/orders/"+repository.FirstOrderID.String(),
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	reqID := res.Header().Get("X-Request-Id")
	assert.NotEmpty(t, reqID)

	parsed, err := uuid.Parse(reqID)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, parsed)
}
