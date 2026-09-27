package http

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProductHandlerTest(t *testing.T) (*httptest.Server, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductRepository(mock)
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, mock
}

func setupProductMuxHandlerTest(t *testing.T) (*chi.Mux, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductRepository(mock)
	svc := service.NewProductService(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	return r, mock
}

func TestGetProducts_WhenProductsExist_Returns200(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now()).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now()),
		)

	res, err := http.Get(srv.URL + "/products")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var resProducts []map[string]any
	err = json.Unmarshal(body, &resProducts)
	assert.NoError(t, err)

	expectedProducts := []map[string]any{
		{
			"id":          FirstUUID.String(),
			"name":        "MacBook Pro",
			"category":    "ACCESSORY",
			"description": "Apple laptop",
			"price":       2500.0,
		},
		{
			"id":          SecondUUID.String(),
			"name":        "iPhone",
			"category":    "ACCESSORY",
			"description": "Apple smartphone",
			"price":       1200.0,
		},
		{
			"id":          ThirdUUID.String(),
			"name":        "hoodie Mykonos",
			"category":    "CLOTHES",
			"description": "Comfortable hoodie",
			"price":       80.0,
		},
		{
			"id":          FourthUUID.String(),
			"name":        "Eye necklace",
			"category":    "JEWELRY",
			"description": "Stylish eye necklace",
			"price":       150.0,
		},
	}

	// remove createdAt field from the comparison
	for _, product := range resProducts {
		delete(product, "created_at")
	}

	assert.ElementsMatch(t, expectedProducts, resProducts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProducts_WhenDbError_Returns500(t *testing.T) {
	r, mock := setupProductMuxHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnError(errors.New("database unavailable"))

	req := httptest.NewRequest(
		http.MethodGet,
		"/products",
		nil,
	)
	req = req.WithContext(context.WithValue(req.Context(), "errorEnabler", "database unavailable"))
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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProduct_WhenProductExists_Returns200(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()),
		)

	res, err := http.Get(srv.URL + "/products/" + FirstUUID.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var updated product.Product
	err = json.Unmarshal(body, &updated)
	assert.NoError(t, err)
	assert.Equal(t, FirstUUID, updated.ID)
	assert.Equal(t, "MacBook Pro", updated.Name)
	assert.Equal(t, "ACCESSORY", updated.Category)
	assert.Equal(t, new("Apple laptop"), updated.Description)
	assert.Equal(t, 2500.0, updated.Price)
	assert.False(t, updated.CreatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProduct_WhenProductNotExists_Returns404(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{
			"product_id", "name", "category", "description", "price", "created_at",
		}))

	res, err := http.Get(srv.URL + "/products/" + id.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_NOT_FOUND",
			"message": "Product with id [`+id.String()+`] was not found"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProduct_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/1234")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_UUID",
			"message": "invalid UUID"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenOnlyQueryProvided_ReturnsMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now()).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now()),
		)

	res, err := http.Get(srv.URL + "/products/search?query=hoodie")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, ThirdUUID.String(), products[0]["id"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_ReturnsCombinedMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now()).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now()),
		)

	res, err := http.Get(srv.URL + "/products/search?query=necklace&maxPrice=200.0")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, FourthUUID.String(), products[0]["id"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenOnlyCategoryProvided_ReturnsCategoryMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now()).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now()),
		)

	res, err := http.Get(srv.URL + "/products/search?category=accessory")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenAllCriteriaProvided_ReturnsCombinedMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now()).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now()),
		)

	res, err := http.Get(srv.URL + "/products/search?query=hoodie&maxPrice=100.0&category=clothes")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var products []map[string]any
	err = json.Unmarshal(body, &products)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, ThirdUUID.String(), products[0]["id"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenMaxPriceInvalid_Returns400(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)

	res, err := http.Get(srv.URL + "/products/search?maxPrice=abc")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "maxPrice must be a valid number."
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenDbError_Returns500(t *testing.T) {
	r, mock := setupProductMuxHandlerTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnError(errors.New("database unavailable"))

	req := httptest.NewRequest(
		http.MethodGet,
		"/products/search?query=hoodie",
		nil,
	)
	req = req.WithContext(context.WithValue(req.Context(), "errorEnabler", "database unavailable"))
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
	assert.NoError(t, mock.ExpectationsWereMet())
}

var (
	FirstUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001")
	SecondUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002")
	ThirdUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d003")
	FourthUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d004")

	ErrornousID   = "01a0b072-db8f-742a-a289-0e290e1fb901"
	ErrornousUUID = uuid.MustParse(ErrornousID)
)
