package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (*AdminService, *repository.InMemoryProductRepository) {
	t.Helper()
	repo := repository.NewInMemoryProductRepository()
	productService := NewProductService(repo)
	svc := NewAdminService(productService, repo)

	return svc, repo
}

func TestCreateProduct_WhenProductNotExists(t *testing.T) {
	svc, repo := setup(t)

	svc.CreateProduct("5", "MacBook Pro M4", 2501.0)
	p, exists := repo.FindByID("5")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "5",
		Name:  "MacBook Pro M4",
		Price: 2501.0,
	}, p)
}

// TODO maybe fix this behavior?
func TestCreateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	p, exists := repo.FindByID("1")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "1",
		Name:  "MacBook Pro",
		Price: 2500.0,
	}, p)

	svc.CreateProduct("1", "MacBook Pro M4", 2501.0)
	p, exists = repo.FindByID("1")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "1",
		Name:  "MacBook Pro M4",
		Price: 2501.0,
	}, p)
}

// TODO maybe fix this behavior?
func TestUpdateProduct_WhenProductNotExists_CreatesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product does not exist
	_, exists := repo.FindByID("10")

	assert.False(t, exists)

	err := svc.UpdateProduct("10", "whatever", 1201.0)
	p, exists := repo.FindByID("10")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.False(t, exists)
	assert.Empty(t, p)
}

func TestUpdateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	p, exists := repo.FindByID("2")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "2",
		Name:  "iPhone",
		Price: 1200.0,
	}, p)

	err := svc.UpdateProduct("2", "iPhone 7", 1201.0)
	p, exists = repo.FindByID("2")

	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "2",
		Name:  "iPhone 7",
		Price: 1201.0,
	}, p)
}

func TestDeleteProduct_WhenProductNotExists_DoesNotFail(t *testing.T) {
	svc, repo := setup(t)

	// product does not exist
	_, exists := repo.FindByID("11")

	assert.False(t, exists)

	svc.DeleteProduct("11")
	_, exists = repo.FindByID("11")

	assert.False(t, exists)

	products := repo.FindAll()
	assert.Len(t, products, 2)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	svc, repo := setup(t)

	// product exists
	_, exists := repo.FindByID("2")

	assert.True(t, exists)

	svc.DeleteProduct("2")
	_, exists = repo.FindByID("2")

	assert.False(t, exists)

	products := repo.FindAll()
	assert.Len(t, products, 1)
}
