package service

import (
	"commerce-platform/services/products/internal/repository"
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setupProductCategoryServiceTest(t *testing.T) (*ProductCategoryService, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductCategoryRepository(mock)
	svc := NewProductCategoryService(repo)

	return svc, mock
}

func TestProductCategoryService_Validate_WhenCategoryExists_ReturnsNil(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories WHERE name ILIKE $1`).
		WithArgs("ACCESSORY").
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}).
				AddRow("ACCESSORY"),
		)

	normalized, err := svc.Validate(context.Background(), "accessory")

	assert.NoError(t, err)
	assert.Equal(t, "ACCESSORY", normalized)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductCategoryService_Validate_WhenCategoryNotExists_ReturnsInvalidCategory(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories WHERE name ILIKE $1`).
		WithArgs("UNKNOWN").
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}).
				RowError(1, pgx.ErrNoRows),
		)

	normalized, err := svc.Validate(context.Background(), "UNKNOWN")

	assert.Empty(t, normalized)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductCategoryService_Validate_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories WHERE name ILIKE $1`).
		WithArgs("ERRORNOUS_CATEGORY").
		WillReturnError(errors.New("database unavailable"))

	normalized, err := svc.Validate(context.Background(), "errornous_category")

	assert.Empty(t, normalized)
	assert.Error(t, err)
	assert.EqualError(t, err, "validate product category: product category exists: database unavailable")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductCategoryService_GetProductCategories_ReturnsAllCategories(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}).
				AddRow("MAGNET").
				AddRow("POSTCARD").
				AddRow("ACCESSORY").
				AddRow("JEWELRY").
				AddRow("CLOTHES"),
		)

	categories, err := svc.GetProductCategories(context.Background())

	assert.NoError(t, err)
	assert.Len(t, categories, 5)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductCategoryService_GetProductCategories_ReturnsDbError(t *testing.T) {
	svc, mock := setupProductCategoryServiceTest(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnError(errors.New("database unavailable"))

	categories, err := svc.GetProductCategories(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "get product categories: query product categories: database unavailable")
	assert.Nil(t, categories)
	assert.NoError(t, mock.ExpectationsWereMet())
}
