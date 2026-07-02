package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
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

	p, err := svc.CreateProduct("MacBook Pro M4", product.ProductCategory("ACCESSORY"), 2501.0)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	p, exists := repo.FindByID(p.ID)

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:       p.ID,
		Name:     "MacBook Pro M4",
		Category: product.ProductCategory("ACCESSORY"),
		Price:    2501.0,
	}, p)
}

func TestCreateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, repo := setup(t)

	p, err := svc.CreateProduct("MacBook Pro M4", product.ProductCategory("UNKNOWN"), 2501.0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.Empty(t, p)

	products := repo.FindAll()
	assert.Len(t, products, 2)
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	svc, repo := setup(t)
	// product does not exist
	id, _ := uuid.NewV7()

	_, exists := repo.FindByID(id)

	assert.False(t, exists)

	updated, err := svc.UpdateProduct(id, "whatever", product.ProductCategory("ACCESSORY"), 1201.0)
	p, exists := repo.FindByID(id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.False(t, exists)
	assert.Empty(t, p)
	assert.Empty(t, updated)
}

func TestUpdateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	p, exists := repo.FindByID(repository.SecondUUID)

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone",
		Category: product.ProductCategory("ACCESSORY"),
		Price:    1200.0,
	}, p)

	updated, err := svc.UpdateProduct(repository.SecondUUID, "iPhone 7", product.ProductCategory("CLOTHES"), 1201.0)
	p, exists = repo.FindByID(repository.SecondUUID)

	assert.NoError(t, err)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone 7",
		Category: product.ProductCategory("CLOTHES"),
		Price:    1201.0,
	}, updated)
	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone 7",
		Category: product.ProductCategory("CLOTHES"),
		Price:    1201.0,
	}, p)
}

func TestUpdateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, repo := setup(t)

	updated, err := svc.UpdateProduct(repository.SecondUUID, "iPhone 7", product.ProductCategory("UNKNOWN"), 1201.0)
	p, exists := repo.FindByID(repository.SecondUUID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.Empty(t, updated)

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:       repository.SecondUUID,
		Name:     "iPhone",
		Category: product.ProductCategory("ACCESSORY"),
		Price:    1200.0,
	}, p)
}

func TestDeleteProduct_WhenProductNotExists_DoesNotFail(t *testing.T) {
	svc, repo := setup(t)
	id, _ := uuid.NewV7()

	// product does not exist
	_, exists := repo.FindByID(id)

	assert.False(t, exists)

	svc.DeleteProduct(id)
	_, exists = repo.FindByID(id)

	assert.False(t, exists)

	products := repo.FindAll()
	assert.Len(t, products, 2)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	_, exists := repo.FindByID(repository.SecondUUID)

	assert.True(t, exists)

	svc.DeleteProduct(repository.SecondUUID)
	_, exists = repo.FindByID(repository.SecondUUID)

	assert.False(t, exists)

	products := repo.FindAll()
	assert.Len(t, products, 1)
}
