package http

import (
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func productCategoryHandlerCompositionTest(t *testing.T) (*chi.Mux, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductCategoryRepository(mock)
	svc := service.NewProductCategoryServiceImpl(repo)
	handler := NewProductCategoryHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, mock
}

func TestGetProductCategoriesComposition_WhenCategoriesExist_Returns200(t *testing.T) {
	r, mock := productCategoryHandlerCompositionTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}).
				AddRow("MAGNET").
				AddRow("POSTCARD").
				AddRow("ACCESSORY").
				AddRow("JEWELRY").
				AddRow("CLOTHES"),
		)

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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductCategoriesComposition_WhenCategoriesNotExist_Returns200(t *testing.T) {
	r, mock := productCategoryHandlerCompositionTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnRows(pgxmock.NewRows([]string{"name"}))

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
	assert.NoError(t, mock.ExpectationsWereMet())
}
