package http

import (
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func ProductCategoryHandlerTest(t *testing.T) (*chi.Mux, *repository.InMemoryProductCategoryRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := service.NewProductCategoryService(repo)
	handler := NewProductCategoryHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, repo
}

func TestGetProductCategories_WhenCategoriesExist_Returns200(t *testing.T) {
	r, _ := ProductCategoryHandlerTest(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/categories",
		nil,
	)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))

	var resCategories []string
	err := json.Unmarshal(res.Body.Bytes(), &resCategories)
	assert.NoError(t, err)

	expectedCategories := []string{"MAGNET", "POSTCARD", "ACCESSORY", "JEWELRY", "CLOTHES"}
	assert.ElementsMatch(t, expectedCategories, resCategories)
}

func TestGetProductCategories_WhenDbError_Returns500(t *testing.T) {
	r, _ := ProductCategoryHandlerTest(t)

	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/categories",
		nil,
	)
	req = req.WithContext(ctxWithError)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "application/json", res.Header().Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		res.Body.String(),
	)
}
