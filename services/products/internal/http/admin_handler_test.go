package http

import (
	"bytes"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"context"
	"encoding/json"
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

func setupAdminHandlerTest(t *testing.T) (*httptest.Server, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	productService := service.NewProductService(repo)
	categoryRepo := repository.NewInMemoryProductCategoryRepository()
	categoryService := service.NewProductCategoryService(categoryRepo)
	adminSvc := service.NewAdminService(productService, categoryService, repo)
	handler := NewAdminHandler(adminSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, repo
}

func TestGetAdmin_Returns200(t *testing.T) {
	srv, _ := setupAdminHandlerTest(t)

	res, err := http.Get(srv.URL + "/admin")

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "admin", string(body))
}

func TestCreateProduct_WhenRequestValid_CreatesProduct(t *testing.T) {
	tests := []struct {
		testName    string
		name        string
		category    string
		description string
		price       float64
	}{
		{"valid-product", "iPad", "ACCESSORY", "some-description", 999},
		{"empty-description", "Hoodie", "CLOTHES", "", 49},
		{"decimal-price", "Necklace", "JEWELRY", "some-description", 15.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, repo := setupAdminHandlerTest(t)

			reqBody := fmt.Sprintf(
				`{"name":%q,"category":%q,"description":%q,"price":%v}`,
				tt.name, tt.category, tt.description, tt.price,
			)

			res, err := http.Post(
				srv.URL+"/admin/products",
				"application/json",
				bytes.NewBufferString(reqBody),
			)

			assert.NoError(t, err)
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			assert.Equal(t, http.StatusCreated, res.StatusCode)

			// decode response to get the server-assigned ID
			var created product.Product
			err = json.Unmarshal(body, &created)
			assert.NoError(t, err)
			assert.NotEmpty(t, created.ID)
			assert.Equal(t, tt.name, created.Name)
			assert.Equal(t, tt.category, created.Category)
			assert.Equal(t, tt.description, created.Description)
			assert.Equal(t, tt.price, created.Price)
			assert.WithinDuration(t, time.Now(), created.CreatedAt, 2*time.Second)
			assert.Equal(t, "/products/"+created.ID.String(), res.Header.Get("Location"))
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

			// verify it was actually persisted
			p, err := repo.FindByID(context.Background(), created.ID)
			assert.NoError(t, err)
			assert.Equal(t, created.ID, p.ID)
			assert.Equal(t, created.Name, p.Name)
			assert.Equal(t, created.Category, p.Category)
			assert.Equal(t, created.Description, p.Description)
			assert.Equal(t, created.Price, p.Price)
		})
	}
}

func TestCreateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, _ := setupAdminHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/admin/products",
		"application/json",
		bytes.NewBufferString(`{
			"error-to-cause": "extra comma, so invalid json",
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_PRODUCT",
			"message": "invalid product"
		}`,
		string(body),
	)
}

func TestCreateProduct_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, _ := setupAdminHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/admin/products",
		"application/json",
		bytes.NewBufferString(`{
			"id": "",
			"name": "",
			"category": "",
			"price": 0
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "name cannot be blank.; category cannot be blank.; price must be > 0."
		}`,
		string(body),
	)
}

func TestCreateProduct_WhenCategoryInvalid_Returns400(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)

	res, err := http.Post(
		srv.URL+"/admin/products",
		"application/json",
		bytes.NewBufferString(`{
			"name": "iPad",
			"category": "UNKNOWN",
			"price": 999
		}`),
	)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_CATEGORY",
			"message": "invalid category"
		}`,
		string(body),
	)

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 4)
}

func TestUpdateProduct_WhenRequestValid_UpdatesProduct(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+repository.SecondUUID.String(),
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"category": "CLOTHES",
			"description": "Updated description",
			"price": 1500
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	var updated product.Product
	err = json.Unmarshal(body, &updated)
	assert.NoError(t, err)
	assert.Equal(t, repository.SecondUUID, updated.ID)
	assert.Equal(t, "iPhone 15", updated.Name)
	assert.Equal(t, "CLOTHES", updated.Category)
	assert.Equal(t, "Updated description", updated.Description)
	assert.Equal(t, 1500.0, updated.Price)
	assert.False(t, updated.CreatedAt.IsZero())

	p, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)
	assert.Equal(t, repository.SecondUUID, p.ID)
	assert.Equal(t, "iPhone 15", p.Name)
	assert.Equal(t, "CLOTHES", p.Category)
	assert.Equal(t, "Updated description", p.Description)
	assert.Equal(t, 1500.0, p.Price)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
}

func TestUpdateProduct_WhenBadUUID_Returns400(t *testing.T) {
	srv, _ := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/1234",
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"category": "CLOTHES",
			"price": 1500
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

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
}

func TestUpdateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, _ := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+repository.SecondUUID.String(),
		bytes.NewBufferString(`{
			"error-to-cause": "extra comma, so invalid json",
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_PRODUCT",
			"message": "invalid product"
		}`,
		string(body),
	)
}

func TestUpdateProduct_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+repository.FirstUUID.String(),
		bytes.NewBufferString(`{
			"name": "",
			"category": "",
			"price": 0
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "VALIDATION_ERROR",
			"message": "name cannot be blank.; category cannot be blank.; price must be > 0."
		}`,
		string(body),
	)

	p, err := repo.FindByID(context.Background(), repository.FirstUUID)
	assert.NoError(t, err)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, "ACCESSORY", p.Category)
	assert.Equal(t, 2500.0, p.Price)
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)
	id, _ := uuid.NewV7()

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+id.String(),
		bytes.NewBufferString(`{
			"name": "non-existing-product",
			"category": "ACCESSORY",
			"price": 1000.1
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

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

	p, err := repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Empty(t, p)
}

func TestUpdateProduct_WhenCategoryInvalid_Returns400(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+repository.SecondUUID.String(),
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"category": "UNKNOWN",
			"price": 1500
		}`),
	)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "INVALID_CATEGORY",
			"message": "invalid category"
		}`,
		string(body),
	)

	p, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:          repository.SecondUUID,
		Name:        "iPhone",
		Category:    "ACCESSORY",
		Description: "Apple smartphone",
		Price:       1200.0,
		CreatedAt:   p.CreatedAt,
	}, p)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/admin/products/"+repository.SecondUUID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusNoContent, res.StatusCode)

	_, err = repo.FindByID(context.Background(), repository.SecondUUID)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Empty(t, res.Header.Get("Content-Type"))
	assert.Empty(t, res.Body)
}

func TestDeleteProduct_WhenBadUUID_Returns400(t *testing.T) {
	srv, repo := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/admin/products/1234",
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

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

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 4)
}
