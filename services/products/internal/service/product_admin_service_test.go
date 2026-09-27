package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (*AdminService, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductRepository(mock)
	productService := NewProductService(repo)
	categoryRepo := repository.NewInMemoryProductCategoryRepository()
	categoryService := NewProductCategoryService(categoryRepo)
	svc := NewAdminService(productService, categoryService, repo)

	return svc, mock
}

func TestCreateProduct_WhenProductNotExists(t *testing.T) {
	svc, mock := setup(t)
	p := product.Product{
		Name:        "MacBook Pro M4",
		Category:    "ACCESSORY",
		Description: new("some-description"),
		Price:       2501.0,
	}
	mock.ExpectQuery(`
			INSERT INTO products (product_id, name, category, description, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(pgxmock.OfType[uuid.UUID](), p.Name, p.Category, p.Description, p.Price).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).AddRow(uuid.New(), p.Name, p.Category, p.Description, p.Price, time.Now()),
		)

	p, err := svc.CreateProduct(context.Background(), p.Name, p.Category, p.Description, p.Price)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, mock := setup(t)

	p, err := svc.CreateProduct(context.Background(), "MacBook Pro M4", "UNKNOWN", new("some-description"), 2501.0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.Empty(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setup(t)
	p := product.Product{
		Name:        "MacBook Pro M4",
		Category:    "ACCESSORY",
		Description: new("some-description"),
		Price:       2501.0,
	}
	mock.ExpectQuery(`
			INSERT INTO products (product_id, name, category, description, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(pgxmock.OfType[uuid.UUID](), p.Name, p.Category, p.Description, p.Price).
		WillReturnError(errors.New("database unavailable"))

	p, err := svc.CreateProduct(context.Background(), p.Name, p.Category, p.Description, p.Price)

	assert.Error(t, err)
	assert.EqualError(t, err, "create product: save product: database unavailable")
	assert.Empty(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	svc, mock := setup(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}))

	updated, err := svc.UpdateProduct(context.Background(), id, "whatever", "ACCESSORY", new(""), 1201.0)

	var notFoundErr *ErrProductNotFound
	assert.Error(t, err)
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.ProductID)
	assert.Empty(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	svc, mock := setup(t)
	p := &product.Product{
		ID:          SecondUUID,
		Name:        "iPhone 15",
		Category:    "CLOTHES",
		Description: new("Updated description"),
		Price:       1500.0,
		CreatedAt:   time.Now(),
	}
	// product exists
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

	updated, err := svc.UpdateProduct(context.Background(), p.ID, p.Name, p.Category, p.Description, p.Price)

	assert.NoError(t, err)
	assert.Equal(t, p.ID, updated.ID)
	assert.Equal(t, p.Name, updated.Name)
	assert.Equal(t, p.Category, updated.Category)
	assert.Equal(t, p.Description, updated.Description)
	assert.Equal(t, p.Price, updated.Price)
	assert.True(t, updated.CreatedAt.Equal(createdAt))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, mock := setup(t)
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

	updated, err := svc.UpdateProduct(context.Background(), SecondUUID, "iPhone 7", "UNKNOWN", new(""), 1201.0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.Empty(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setup(t)
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
	mock.ExpectQuery(`
			UPDATE products 
			SET name = $1, category = $2, description = $3, price = $4
			WHERE product_id = $5
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs("iPhone 7", "CLOTHES", new(""), 1201.0, SecondUUID).
		WillReturnError(errors.New("database unavailable"))

	updated, err := svc.UpdateProduct(context.Background(), SecondUUID, "iPhone 7", "CLOTHES", new(""), 1201.0)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating product: update product: database unavailable")
	assert.Empty(t, updated)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_WhenProductNotExists_DoesNotFail(t *testing.T) {
	svc, mock := setup(t)
	id, _ := uuid.NewV7()
	mock.ExpectExec(`
			DELETE FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := svc.DeleteProduct(context.Background(), id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	svc, mock := setup(t)
	mock.ExpectExec(`
			DELETE FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := svc.DeleteProduct(context.Background(), SecondUUID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setup(t)
	mock.ExpectExec(`
			DELETE FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnError(errors.New("database unavailable"))

	err := svc.DeleteProduct(context.Background(), SecondUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "deleting product: delete product: database unavailable")
	assert.NoError(t, mock.ExpectationsWereMet())
}
