package http

import (
	"bytes"
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/service"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupAdminHandlerTest(t *testing.T) (*httptest.Server, *fakeProductService) {
	t.Helper()
	adminSvc := newFakeProductService(t)
	handler := NewAdminHandler(adminSvc)

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	return srv, adminSvc
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
			mock.expectCreateProduct(
				func(ctx context.Context, name, category string, description *string, price float64) (product.Product, error) {
					return product.Product{
						ID:          uuid.New(),
						Name:        tt.name,
						Category:    tt.category,
						Description: new(tt.description),
						Price:       tt.price,
						CreatedAt:   time.Now()}, nil
				})

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
			assert.Equal(t, tt.name, mock.created[0].Name)
			assert.Equal(t, tt.category, mock.created[0].Category)
			assert.Equal(t, tt.description, *mock.created[0].Description)
			assert.Equal(t, tt.price, mock.created[0].Price)
		})
	}
}

func TestCreateProduct_WhenRequestBodyTooBig_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

	reqBody := fmt.Sprintf(
		`{"name":%q,"category":%q,"description":%q,"price":%v}`,
		strings.Repeat("a", 1<<10), "category", "description", 100,
	)

	res, err := http.Post(
		srv.URL+"/admin/products",
		"application/json",
		bytes.NewBufferString(reqBody),
	)
	assert.NoError(t, err)
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "REQUEST_BODY_TOO_LONG",
			"message": "Request body exceeds 1024 bytes"
		}`,
		string(body),
	)
	assert.Empty(t, mock.created)
}

func TestCreateProduct_WhenRequestBodyWithTrailingJson_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

	reqBody := fmt.Sprintf(
		`{"name":%q,"category":%q,"description":%q,"price":%v}{"trailing": "json"}`,
		"name", "category", "description", 100,
	)

	res, err := http.Post(
		srv.URL+"/admin/products",
		"application/json",
		bytes.NewBufferString(reqBody),
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
	assert.Empty(t, mock.created)
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
	assert.Empty(t, mock.created)
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
	assert.Empty(t, mock.created)
}

func TestCreateProduct_WhenCategoryInvalid_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	mock.expectCreateProduct(
		func(ctx context.Context, name, category string, description *string, price float64) (product.Product, error) {
			return product.Product{}, service.ErrInvalidCategory
		})

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
	assert.Equal(t, "iPad", mock.created[0].Name)
	assert.Equal(t, "UNKNOWN", mock.created[0].Category)
	assert.Nil(t, mock.created[0].Description)
	assert.Equal(t, 999.0, mock.created[0].Price)
}

func TestUpdateProduct_WhenRequestValid_UpdatesProduct(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	p := product.Product{
		ID:          SecondUUID,
		Name:        "iPhone 15",
		Category:    "CLOTHES",
		Description: new("Updated description"),
		Price:       1500.0,
		CreatedAt:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	mock.expectUpdateProduct(
		func(ctx context.Context, id uuid.UUID, name, category string, description *string, price float64) (product.Product, error) {
			return p, nil
		})

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
	assert.Equal(t, p, updated)
	assert.Equal(t, p.ID, mock.updated[0].ID)
	assert.Equal(t, p.Name, mock.updated[0].Name)
	assert.Equal(t, p.Category, mock.updated[0].Category)
	assert.Equal(t, p.Description, mock.updated[0].Description)
	assert.Equal(t, p.Price, mock.updated[0].Price)
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
	assert.Empty(t, mock.updated)
}

func TestUpdateProduct_WhenRequestBodyTooBig_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+SecondUUID.String(),
		bytes.NewBufferString(`{
			"name": "`+strings.Repeat("a", 1<<10)+`",
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

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "REQUEST_BODY_TOO_LONG",
			"message": "Request body exceeds 1024 bytes"
		}`,
		string(body),
	)
	assert.Empty(t, mock.updated)
}

func TestUpdateProduct_WhenRequestBodyWithTrailingJson_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)

	req, err := http.NewRequest(
		http.MethodPut,
		srv.URL+"/admin/products/"+SecondUUID.String(),
		bytes.NewBufferString(`{
			"name": "iPhone 15",
			"category": "CLOTHES",
			"description": "Updated description",
			"price": 1500
		}{"trailing": "json"}`),
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
	assert.Empty(t, mock.updated)
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
	assert.Empty(t, mock.updated)
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
	assert.Empty(t, mock.updated)
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.expectUpdateProduct(
		func(ctx context.Context, id uuid.UUID, name, category string, description *string, price float64) (product.Product, error) {
			return product.Product{}, &service.ErrProductNotFound{ProductID: id}
		})

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
	assert.Equal(t, id, mock.updated[0].ID)
	assert.Equal(t, "non-existing-product", mock.updated[0].Name)
	assert.Equal(t, "ACCESSORY", mock.updated[0].Category)
	assert.Nil(t, mock.updated[0].Description)
	assert.Equal(t, 1000.1, mock.updated[0].Price)
}

func TestUpdateProduct_WhenCategoryInvalid_Returns400(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	mock.expectUpdateProduct(
		func(ctx context.Context, id uuid.UUID, name, category string, description *string, price float64) (product.Product, error) {
			return product.Product{}, service.ErrInvalidCategory
		})

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
	assert.Equal(t, SecondUUID, mock.updated[0].ID)
	assert.Equal(t, "iPhone 15", mock.updated[0].Name)
	assert.Equal(t, "UNKNOWN", mock.updated[0].Category)
	assert.Nil(t, mock.updated[0].Description)
	assert.Equal(t, 1500.0, mock.updated[0].Price)
}

func TestUpdateProduct_WhenUpdateConflict_Returns409(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	id, _ := uuid.NewV7()
	mock.expectUpdateProduct(
		func(ctx context.Context, id uuid.UUID, name, category string, description *string, price float64) (product.Product, error) {
			return product.Product{}, service.ErrProductUpdateConflict
		})

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

	assert.Equal(t, http.StatusConflict, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
	assert.JSONEq(
		t,
		`{
			"code": "PRODUCT_UPDATE_CONFLICT",
			"message": "product was updated by another process"
		}`,
		string(body),
	)
	assert.Equal(t, id, mock.updated[0].ID)
	assert.Equal(t, "non-existing-product", mock.updated[0].Name)
	assert.Equal(t, "ACCESSORY", mock.updated[0].Category)
	assert.Nil(t, mock.updated[0].Description)
	assert.Equal(t, 1000.1, mock.updated[0].Price)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	mock.expectDeleteProduct(
		func(ctx context.Context, u uuid.UUID) error {
			return nil
		})

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
	assert.Equal(t, SecondUUID, mock.deleted[0])
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
	assert.Empty(t, mock.deleted)
}

func TestDeleteProduct_WhenDbError_Returns500(t *testing.T) {
	srv, mock := setupAdminHandlerTest(t)
	mock.expectDeleteProduct(
		func(ctx context.Context, u uuid.UUID) error {
			return errors.New("unexpected error")
		})

	req, err := http.NewRequest(
		http.MethodDelete,
		srv.URL+"/admin/products/"+SecondUUID.String(),
		nil,
	)
	assert.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

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
	assert.Equal(t, SecondUUID, mock.deleted[0])
}
