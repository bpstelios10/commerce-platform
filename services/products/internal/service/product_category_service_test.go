package service

import (
	"commerce-platform/services/products/internal/repository"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductCategoryService_Validate_WhenCategoryExists_ReturnsNil(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	normalized, err := svc.Validate(context.Background(), "accessory")

	assert.NoError(t, err)
	assert.Equal(t, "ACCESSORY", normalized)
}

func TestProductCategoryService_Validate_WhenCategoryNotExists_ReturnsInvalidCategory(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	normalized, err := svc.Validate(context.Background(), "UNKNOWN")

	assert.Empty(t, normalized)
	assert.ErrorIs(t, err, ErrInvalidCategory)
}

func TestProductCategoryService_Validate_WhenDbError_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	normalized, err := svc.Validate(context.Background(), "errornous_category")

	assert.Empty(t, normalized)
	assert.Error(t, err)
	assert.EqualError(t, err, "validate product category: unexpected error")
}

func TestProductCategoryService_GetProductCategories_ReturnsAllCategories(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	categories, err := svc.GetProductCategories(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, categories)
}

func TestProductCategoryService_GetProductCategories_ReturnsDbError(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	categories, err := svc.GetProductCategories(ctxWithError)

	assert.Error(t, err)
	assert.EqualError(t, err, "get product categories: unexpected error")
	assert.Nil(t, categories)
}
