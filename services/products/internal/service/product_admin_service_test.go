package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateProduct_WhenProductNotExists(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	adminSvc := NewAdminService(repo)

	adminSvc.CreateProduct("5", "MacBook Pro M4", 2501.0)
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
	repo := repository.NewInMemoryProductRepository()
	adminSvc := NewAdminService(repo)

	// product exists
	p, exists := repo.FindByID("1")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "1",
		Name:  "MacBook Pro",
		Price: 2500.0,
	}, p)

	adminSvc.CreateProduct("1", "MacBook Pro M4", 2501.0)
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
	repo := repository.NewInMemoryProductRepository()
	adminSvc := NewAdminService(repo)

	// product does not exist
	_, exists := repo.FindByID("10")

	assert.False(t, exists)

	adminSvc.UpdateProduct("10", "whatever", 1201.0)
	p, exists := repo.FindByID("10")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "10",
		Name:  "whatever",
		Price: 1201.0,
	}, p)
}

func TestUpdateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	adminSvc := NewAdminService(repo)

	// product exists
	p, exists := repo.FindByID("2")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "2",
		Name:  "iPhone",
		Price: 1200.0,
	}, p)

	adminSvc.UpdateProduct("2", "iPhone 7", 1201.0)
	p, exists = repo.FindByID("2")

	assert.True(t, exists)
	assert.Equal(t, product.Product{
		ID:    "2",
		Name:  "iPhone 7",
		Price: 1201.0,
	}, p)
}

func TestDeleteProduct_WhenProductNotExists_NothingChanges(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	adminSvc := NewAdminService(repo)

	// product does not exist
	_, exists := repo.FindByID("11")

	assert.False(t, exists)

	adminSvc.DeleteProduct("11")
	_, exists = repo.FindByID("11")

	assert.False(t, exists)

	products := repo.FindAll()
	assert.Len(t, products, 2)
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	adminSvc := NewAdminService(repo)

	// product exists
	_, exists := repo.FindByID("2")

	assert.True(t, exists)

	adminSvc.DeleteProduct("2")
	_, exists = repo.FindByID("2")

	assert.False(t, exists)

	products := repo.FindAll()
	assert.Len(t, products, 1)
}
