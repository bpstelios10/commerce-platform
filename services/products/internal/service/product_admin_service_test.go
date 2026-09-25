package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (*AdminService, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	productService := NewProductService(repo)
	categoryRepo := repository.NewInMemoryProductCategoryRepository()
	categoryService := NewProductCategoryService(categoryRepo)
	svc := NewAdminService(productService, categoryService, repo)

	return svc, repo
}

func TestCreateProduct_WhenProductNotExists(t *testing.T) {
	svc, repo := setup(t)

	p, err := svc.CreateProduct(context.Background(), "MacBook Pro M4", "ACCESSORY", 2501.0, 10)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	p, err = repo.FindByID(context.Background(), p.ID)
	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       p.ID,
		Name:     "MacBook Pro M4",
		Category: "ACCESSORY",
		Price:    2501.0,
		Stock:    10,
	}, p)
}

func TestCreateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, repo := setup(t)

	p, err := svc.CreateProduct(context.Background(), "MacBook Pro M4", "UNKNOWN", 2501.0, 10)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.Empty(t, p)

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 4)
}

func TestCreateProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, repo := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	p, err := svc.CreateProduct(ctxWithError, "MacBook Pro M4", "ACCESSORY", 2501.0, 10)

	assert.Error(t, err)
	assert.EqualError(t, err, "create product: unexpected error")
	assert.Empty(t, p)

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 4)
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	svc, repo := setup(t)
	// product does not exist
	id, _ := uuid.NewV7()

	_, err := repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	updated, err := svc.UpdateProduct(context.Background(), id, "whatever", "ACCESSORY", 1201.0, 10)

	var notFoundErr *ErrProductNotFound
	assert.Error(t, err)
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.ProductID)
	assert.Empty(t, updated)

	p, err := repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Empty(t, p)
}

func TestUpdateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	p, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone",
		Category: "ACCESSORY",
		Price:    1200.0,
		Stock:    5,
	}, p)

	updated, err := svc.UpdateProduct(context.Background(), repository.SecondUUID, "iPhone 7", "CLOTHES", 1201.0, 11)

	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone 7",
		Category: "CLOTHES",
		Price:    1201.0,
		Stock:    11,
	}, updated)

	p, err = repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone 7",
		Category: "CLOTHES",
		Price:    1201.0,
		Stock:    11,
	}, p)
}

func TestUpdateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, repo := setup(t)

	updated, err := svc.UpdateProduct(context.Background(), repository.SecondUUID, "iPhone 7", "UNKNOWN", 1201.0, 11)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.Empty(t, updated)

	p, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone",
		Category: "ACCESSORY",
		Price:    1200.0,
		Stock:    5,
	}, p)
}

func TestUpdateProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, repo := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	updated, err := svc.UpdateProduct(ctxWithError, repository.SecondUUID, "iPhone 7", "CLOTHES", 1201.0, 11)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating product: unexpected error")
	assert.Empty(t, updated)

	p, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone",
		Category: "ACCESSORY",
		Price:    1200.0,
		Stock:    5,
	}, p)
}

func TestDeleteProduct_WhenProductNotExists_DoesNotFail(t *testing.T) {
	svc, repo := setup(t)
	id, _ := uuid.NewV7()

	// product does not exist
	_, err := repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	svc.DeleteProduct(context.Background(), id)
	_, err = repo.FindByID(context.Background(), id)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 4)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	_, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)

	svc.DeleteProduct(context.Background(), repository.SecondUUID)

	_, err = repo.FindByID(context.Background(), repository.SecondUUID)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 3)
}

func TestDeleteProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, repo := setup(t)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	// product exists
	_, err := repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)

	err = svc.DeleteProduct(ctxWithError, repository.SecondUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "deleting product: unexpected error")

	_, err = repo.FindByID(context.Background(), repository.SecondUUID)
	assert.NoError(t, err)

	products, err := repo.FindAll(context.Background())
	assert.NoError(t, err)
	assert.Len(t, products, 4)
}
