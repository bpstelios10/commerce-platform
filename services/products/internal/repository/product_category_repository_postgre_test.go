package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setupCategoriesRepo(t *testing.T) (pgxmock.PgxPoolIface, *PostgreProductCategoryRepository) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})

	return mock, NewPostgreProductCategoryRepository(mock)
}

func TestExists_WhenCategoryExists_ReturnsTrue(t *testing.T) {
	mock, repo := setupCategoriesRepo(t)
	mock.ExpectQuery(`SELECT name FROM product_categories WHERE name ILIKE $1`).
		WithArgs("CATEG").
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}).
				AddRow("CATEG"),
		)

	result, err := repo.Exists(context.Background(), "CATEG")

	assert.NoError(t, err)
	assert.True(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExists_WhenCategoryNotExists_ReturnsFalse(t *testing.T) {
	mock, repo := setupCategoriesRepo(t)
	mock.ExpectQuery(`SELECT name FROM product_categories WHERE name ILIKE $1`).
		WithArgs("CATEG").
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}))

	result, err := repo.Exists(context.Background(), "CATEG")

	assert.NoError(t, err)
	assert.False(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestExists_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setupCategoriesRepo(t)
	mock.ExpectQuery(`SELECT name FROM product_categories WHERE name ILIKE $1`).
		WithArgs("CATEG").
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.Exists(context.Background(), "CATEG")

	assert.Error(t, err)
	assert.False(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_WhenCategoriesExist_ReturnsAllCategories(t *testing.T) {
	mock, repo := setupCategoriesRepo(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnRows(
			pgxmock.NewRows([]string{"name"}).
				AddRow("MAGNET").
				AddRow("POSTCARD").
				AddRow("ACCESSORY").
				AddRow("JEWELRY").
				AddRow("CLOTHES"),
		)

	result, err := repo.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 5)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setupCategoriesRepo(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.GetAll(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "query product categories: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAll_WhenRowWithError_StopsScanning(t *testing.T) {
	mock, repo := setupCategoriesRepo(t)
	mock.ExpectQuery(`SELECT name FROM product_categories`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"name",
			}).
				AddRow("MAGNET").
				RowError(1, errors.New("row error")).
				AddRow("POSTCARD"),
		)

	result, err := repo.GetAll(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "scan product categories: row error")
	assert.Len(t, result, 1)
	assert.Contains(t, result, "MAGNET")
	assert.NoError(t, mock.ExpectationsWereMet())
}
