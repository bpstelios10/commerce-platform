package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductCategoryService_Validate_WhenCategoryExists_ReturnsNil(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	err := svc.Validate(product.ProductCategory("ACCESSORY"))

	assert.NoError(t, err)
}

func TestProductCategoryService_Validate_WhenCategoryDoesNotExist_ReturnsInvalidCategory(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	err := svc.Validate(product.ProductCategory("UNKNOWN"))

	assert.ErrorIs(t, err, ErrInvalidCategory)
}
