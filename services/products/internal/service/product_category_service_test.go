package service

import (
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProductCategoryService_Validate_WhenCategoryExists_ReturnsNil(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	normalized, err := svc.Validate("accessory")

	assert.NoError(t, err)
	assert.Equal(t, "ACCESSORY", normalized)
}

func TestProductCategoryService_Validate_WhenCategoryDoesNotExist_ReturnsInvalidCategory(t *testing.T) {
	repo := repository.NewInMemoryProductCategoryRepository()
	svc := NewProductCategoryService(repo)
	normalized, err := svc.Validate("UNKNOWN")

	assert.Empty(t, normalized)
	assert.ErrorIs(t, err, ErrInvalidCategory)
}
