package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func productCategoryHandlerTest(t *testing.T) (*chi.Mux, *fakeProductCategoryService) {
	t.Helper()
	svc := newFakeProductCategoryService(t)
	handler := NewProductCategoryHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, svc
}

func TestGetProductCategories_WhenCategoriesExist_Returns200(t *testing.T) {
	r, mock := productCategoryHandlerTest(t)
	mock.expectGetProductCategories(
		func(ctx context.Context) ([]string, error) {
			return []string{"MAGNET", "POSTCARD", "ACCESSORY", "JEWELRY", "CLOTHES"}, nil
		})

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

func TestGetProductCategories_WhenCategoriesNotExist_Returns200(t *testing.T) {
	r, mock := productCategoryHandlerTest(t)
	mock.expectGetProductCategories(
		func(ctx context.Context) ([]string, error) {
			return []string{}, nil
		})

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
	assert.Empty(t, resCategories)
}

func TestGetProductCategories_WhenDbError_Returns500(t *testing.T) {
	r, mock := productCategoryHandlerTest(t)
	mock.expectGetProductCategories(
		func(ctx context.Context) ([]string, error) {
			return []string{}, errors.New("unexpected error")
		})

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/categories",
		nil,
	)
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
