package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetProducts_WhenProductExists_ReturnsProducts(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p := svc.GetProducts()

	assert.Equal(t, 2, len(p))
}

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProductByID(repository.FirstUUID)

	assert.NoError(t, err)
	assert.Equal(t, repository.FirstUUID, p.ID)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, 2500.0, p.Price)
}

func TestGetProductByID_WhenProductDoesNotExist_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	id, _ := uuid.NewV7()

	p, err := svc.GetProductByID(id)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Equal(t, product.Product{}, p)
}
