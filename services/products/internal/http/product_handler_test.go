package http

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupProductHandlerTest(t *testing.T) (*httptest.Server, *fakeProductService) {
	t.Helper()
	svc := newFakeProductService(t)
	handler := NewProductHandler(svc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, svc
}

func TestGetProducts_WhenProductsExist_Returns200(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectGetProducts(func(ctx context.Context) ([]product.Product, error) {
		return allProducts, nil
	})

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
}

func TestGetProducts_WhenDbError_Returns500(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectGetProducts(func(ctx context.Context) ([]product.Product, error) {
		return []product.Product{}, errors.New("database unavailable")
	})

	res, err := http.Get(srv.URL + "/products")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		string(body),
	)
}

func TestGetProducts_WhenRequestCancelled_Returns499(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectGetProducts(func(ctx context.Context) ([]product.Product, error) {
		return []product.Product{}, context.Canceled
	})

	res, err := http.Get(srv.URL + "/products")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, 499, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "REQUEST_CANCELLED",
			"message": "request canceled by client"
		}`,
		string(body),
	)
}

func TestGetProduct_WhenProductExists_Returns200(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectGetProductByID(
		func(ctx context.Context, u uuid.UUID) (product.Product, error) {
			return allProducts[0], nil
		})

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
	assert.Equal(t, FirstUUID, mock.searchedIDs[0])
}

func TestGetProduct_WhenProductNotExists_Returns404(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.expectGetProductByID(
		func(ctx context.Context, u uuid.UUID) (product.Product, error) {
			return product.Product{}, fmt.Errorf("blah: %w", &service.ErrProductNotFound{ProductID: id})
		})

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
	assert.Equal(t, id, mock.searchedIDs[0])
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
	assert.Empty(t, mock.searchedIDs)
}

func TestSearchProducts_WhenOnlyQueryProvided_ReturnsMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectSearchProducts(
		func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
			if query == "hoodie" {
				return []product.Product{allProducts[2]}, nil
			}
			return allProducts, nil
		})

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
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_ReturnsCombinedMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectSearchProducts(
		func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
			if query == "necklace" && *maxPrice == 200.0 {
				return []product.Product{allProducts[3]}, nil
			}
			return allProducts, nil
		})

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
}

func TestSearchProducts_WhenOnlyCategoryProvided_ReturnsCategoryMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectSearchProducts(
		func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
			if category == "accessory" {
				return []product.Product{allProducts[0], allProducts[1]}, nil
			}
			return allProducts, nil
		})

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
}

func TestSearchProducts_WhenAllCriteriaProvided_ReturnsCombinedMatches(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectSearchProducts(
		func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
			if query == "hoodie" && *maxPrice == 100.0 && category == "clothes" {
				return []product.Product{allProducts[2]}, nil
			}
			return allProducts, nil
		})

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
}

func TestSearchProducts_WhenMaxPriceInvalid_Returns400(t *testing.T) {
	srv, _ := setupProductHandlerTest(t)

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
			"message": "maxPrice must be a valid number, of format {xxxxxx.xx}."
		}`,
		string(body),
	)
}

func TestSearchProducts_WhenServiceError_Returns500(t *testing.T) {
	srv, mock := setupProductHandlerTest(t)
	mock.expectSearchProducts(
		func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
			return allProducts, errors.New("some error")
		})

	res, err := http.Get(srv.URL + "/products/search?query=hoodie")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		string(body),
	)
}

var (
	FirstUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001")
	SecondUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002")
	ThirdUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d003")
	FourthUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d004")

	allProducts = []product.Product{
		{
			ID:          FirstUUID,
			Name:        "MacBook Pro",
			Category:    "ACCESSORY",
			Description: new("Apple laptop"),
			Price:       2500,
			CreatedAt:   time.Now(),
			Version:     1,
		},
		{
			ID:          SecondUUID,
			Name:        "iPhone",
			Category:    "ACCESSORY",
			Description: new("Apple smartphone"),
			Price:       1200,
			CreatedAt:   time.Now(),
			Version:     1,
		},
		{
			ID:          ThirdUUID,
			Name:        "hoodie Mykonos",
			Category:    "CLOTHES",
			Description: new("Comfortable hoodie"),
			Price:       80,
			CreatedAt:   time.Now(),
			Version:     1,
		},
		{
			ID:          FourthUUID,
			Name:        "Eye necklace",
			Category:    "JEWELRY",
			Description: new("Stylish eye necklace"),
			Price:       150,
			CreatedAt:   time.Now(),
			Version:     1,
		},
	}
)
