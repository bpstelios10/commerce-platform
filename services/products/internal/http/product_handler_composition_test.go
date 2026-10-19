package http

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setupProductHandlerCompositionTest(t *testing.T) (*httptest.Server, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductRepository(mock, repository.Timeouts{
		Read:  100 * time.Millisecond,
		List:  100 * time.Millisecond,
		Write: 100 * time.Millisecond,
	})
	svc := service.NewProductServiceImpl(repo)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, mock
}

func TestGetProductsComposition_WhenProductsExist_Returns200(t *testing.T) {
	srv, mock := setupProductHandlerCompositionTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at, version FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at", "version",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now(), 1).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now(), 1).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now(), 1).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now(), 1),
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
			"version":     float64(1),
		},
		{
			"id":          SecondUUID.String(),
			"name":        "iPhone",
			"category":    "ACCESSORY",
			"description": "Apple smartphone",
			"price":       1200.0,
			"version":     float64(1),
		},
		{
			"id":          ThirdUUID.String(),
			"name":        "hoodie Mykonos",
			"category":    "CLOTHES",
			"description": "Comfortable hoodie",
			"price":       80.0,
			"version":     float64(1),
		},
		{
			"id":          FourthUUID.String(),
			"name":        "Eye necklace",
			"category":    "JEWELRY",
			"description": "Stylish eye necklace",
			"price":       150.0,
			"version":     float64(1),
		},
	}

	// remove createdAt field from the comparison
	for _, product := range resProducts {
		delete(product, "created_at")
	}

	assert.ElementsMatch(t, expectedProducts, resProducts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductComposition_WhenProductExists_Returns200(t *testing.T) {
	srv, mock := setupProductHandlerCompositionTest(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at, version
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at", "version",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now(), 1),
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
	assert.Equal(t, 1, updated.Version)
	assert.False(t, updated.CreatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductComposition_WhenProductNotExists_Returns404(t *testing.T) {
	srv, mock := setupProductHandlerCompositionTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at, version
			FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{
			"product_id", "name", "category", "description", "price", "created_at", "version",
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

func TestGetProductComposition_WhenQueryExceedsTimeout_Returns503(t *testing.T) {
	srv, mock := setupProductHandlerCompositionTest(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at, version
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillDelayFor(120 * time.Millisecond).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at", "version",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now(), 1),
		)

	res, err := http.Get(srv.URL + "/products/" + FirstUUID.String())

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "REQUEST_TIMED_OUT",
			"message": "request timed out"
		}`,
		string(body),
	)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProductsComposition_WhenOnlyQueryProvided_ReturnsMatches(t *testing.T) {
	srv, mock := setupProductHandlerCompositionTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at, version FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at", "version",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now(), 1).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now(), 1).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now(), 1).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now(), 1),
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
	assert.Equal(t, float64(1), products[0]["version"])
	assert.NoError(t, mock.ExpectationsWereMet())
}
