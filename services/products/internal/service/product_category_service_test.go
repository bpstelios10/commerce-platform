package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupProductCategoryServiceTest(t *testing.T) (*ProductCategoryService, *fakeProductCategoryRepo) {
	t.Helper()
	repo := newFakeProductCategoryRepo(t)
	svc := NewProductCategoryService(repo)

	return svc, repo
}

func TestProductCategoryService_Validate_WhenCategoryExists_ReturnsNil(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			if s == "ACCESSORY" {
				return true, nil
			}
			return false, nil
		})

	normalized, err := svc.Validate(context.Background(), "accessory")

	assert.NoError(t, err)
	assert.Equal(t, "ACCESSORY", normalized)
	assert.Equal(t, "ACCESSORY", mock.searchedCategories[0])
}

func TestProductCategoryService_Validate_WhenCategoryNotExists_ReturnsInvalidCategory(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return false, nil
		})

	normalized, err := svc.Validate(context.Background(), "UNKNOWN")

	assert.Empty(t, normalized)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.EqualError(t, err, "validate product category UNKNOWN: invalid category")
	assert.Equal(t, "UNKNOWN", mock.searchedCategories[0])
}

func TestProductCategoryService_Validate_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return false, errors.New("product category exists: database unavailable")
		})

	normalized, err := svc.Validate(context.Background(), "errornous_category")

	assert.Empty(t, normalized)
	assert.Error(t, err)
	assert.EqualError(t, err, "validate product category: product category exists: database unavailable")
}

func TestProductCategoryService_GetProductCategories_ReturnsAllCategories(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.expectGetAll(
		func(ctx context.Context) ([]string, error) {
			return []string{"MAGNET", "POSTCARD", "ACCESSORY", "JEWELRY", "CLOTHES"}, nil
		})

	categories, err := svc.GetProductCategories(context.Background())

	assert.NoError(t, err)
	assert.Len(t, categories, 5)
}

func TestProductCategoryService_GetProductCategories_ReturnsDbError(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.expectGetAll(
		func(ctx context.Context) ([]string, error) {
			return []string{}, errors.New("query product categories: database unavailable")
		})

	categories, err := svc.GetProductCategories(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "get product categories: query product categories: database unavailable")
	assert.Nil(t, categories)
}
