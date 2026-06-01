package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProductByID("1")

	assert.NoError(t, err)
	assert.Equal(t, "1", p.ID)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, 2500.0, p.Price)
}

func TestGetProductByID_WhenProductDoesNotExist_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProductByID("10")

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Equal(t, product.Product{}, p)
}
