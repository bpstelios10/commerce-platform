package http

import (
	"bytes"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"commerce-platform/services/products/internal/service"
	"encoding/json"
	"fmt"
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

func setupAdminHandlerTest(t *testing.T) (*httptest.Server, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductRepository(mock)
	productService := service.NewProductService(repo)
	categoryRepo := repository.NewInMemoryProductCategoryRepository()
	categoryService := service.NewProductCategoryService(categoryRepo)
	adminSvc := service.NewAdminService(productService, categoryService, repo)
	handler := NewAdminHandler(adminSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, mock
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
			srv, mock := setupAdminHandlerTest(t)
			mock.ExpectQuery(`
					INSERT INTO products (product_id, name, category, description, price)
					VALUES ($1, $2, $3, $4, $5)
					RETURNING product_id, name, category, description, price, created_at;`).
				WithArgs(pgxmock.OfType[uuid.UUID](), tt.name, tt.category, new(tt.description), tt.price).
				WillReturnRows(
					pgxmock.NewRows([]string{
						"product_id", "name", "category", "description", "price", "created_at",
					}).AddRow(uuid.New(), tt.name, tt.category, new(tt.description), tt.price, time.Now()),
				)

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
			assert.Equal(t, new(tt.description), created.Description)
			assert.Equal(t, tt.price, created.Price)
			assert.WithinDuration(t, time.Now(), created.CreatedAt, 2*time.Second)
			assert.Equal(t, "/products/"+created.ID.String(), res.Header.Get("Location"))
			assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct_WhenCategoryInvalid_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenRequestValid_UpdatesProduct(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	p := &product.Product{
		ID:          SecondUUID,
		Name:        "iPhone 15",
		Category:    "CLOTHES",
		Description: new("Updated description"),
		Price:       1500.0,
	}
	createdAt := time.Now()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, createdAt),
		)
	mock.ExpectQuery(`
			UPDATE products 
			SET name = $1, category = $2, description = $3, price = $4
			WHERE product_id = $5
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(p.Name, p.Category, p.Description, p.Price, p.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).AddRow(p.ID, p.Name, p.Category, p.Description, p.Price, createdAt),
		)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+SecondUUID.String(),
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
	assert.Equal(t, p.ID, updated.ID)
	assert.Equal(t, p.Name, updated.Name)
	assert.Equal(t, p.Category, updated.Category)
	assert.Equal(t, p.Description, updated.Description)
	assert.Equal(t, p.Price, updated.Price)
	assert.True(t, updated.CreatedAt.Equal(createdAt))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenBadRequestBody_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+SecondUUID.String(),
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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenRequestInvalid_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+FirstUUID.String(),
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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{
			"product_id", "name", "category", "description", "price", "created_at",
		}))

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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenCategoryInvalid_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()),
		)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+SecondUUID.String(),
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
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	mock.ExpectExec(`
			DELETE FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/admin/products/"+SecondUUID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	assert.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_WhenBadUUID_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

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
	assert.NoError(t, mock.ExpectationsWereMet())
}
