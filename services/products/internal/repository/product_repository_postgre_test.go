package repository

import (
	"commerce-platform/services/products/internal/product"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) (pgxmock.PgxPoolIface, *PostgreProductRepository) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})

	return mock, NewPostgreProductRepository(mock)
}

func TestFindAll_WhenProductsExist_ReturnsAllProducts(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()).
				AddRow(SecondUUID, "iPhone", "ACCESSORY", new("Apple smartphone"), 1200, time.Now()).
				AddRow(ThirdUUID, "hoodie Mykonos", "CLOTHES", new("Comfortable hoodie"), 80, time.Now()).
				AddRow(FourthUUID, "Eye necklace", "JEWELRY", new("Stylish eye necklace"), 150, time.Now()),
		)

	result, err := repo.FindAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 4)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "query products: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow("invalid-uuid", "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()),
		)

	result, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "scan product: scanning value error for column 'product_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindAll_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500.0, time.Now()).
				RowError(1, errors.New("read failed")),
		)

	result, err := repo.FindAll(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "iterate products: read failed")
	assert.Len(t, result, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenProductsExist_ReturnsProduct(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price, testP.CreatedAt),
		)

	result, err := repo.FindByID(context.Background(), FirstUUID)

	assert.NoError(t, err)
	assert.Equal(t, testP, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.FindByID(context.Background(), FirstUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "find product by id: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow("invalid-uuid", "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()),
		)

	result, err := repo.FindByID(context.Background(), FirstUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "find product by id: scanning value error for column 'product_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				RowError(1, pgx.ErrNoRows),
		)

	result, err := repo.FindByID(context.Background(), FirstUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "find product by id: not found")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_WhenNoDbIssues_ReturnsNewProduct(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO products (product_id, name, category, description, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price, testP.CreatedAt),
		)

	result, err := repo.Save(context.Background(), testP)

	assert.NoError(t, err)
	assert.Equal(t, testP, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO products (product_id, name, category, description, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.Save(context.Background(), testP)

	assert.Error(t, err)
	assert.EqualError(t, err, "save product: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO products (product_id, name, category, description, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow("invalid-uuid", testP.Name, testP.Category, testP.Description, testP.Price, testP.CreatedAt),
		)

	result, err := repo.Save(context.Background(), testP)

	assert.Error(t, err)
	assert.EqualError(t, err, "save product: scanning value error for column 'product_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSave_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			INSERT INTO products (product_id, name, category, description, price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				RowError(1, errors.New("read failed")),
		)

	result, err := repo.Save(context.Background(), testP)

	assert.Error(t, err)
	assert.EqualError(t, err, "save product: no rows in result set")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenNoDbIssues_ReturnsUpdatedProduct(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE products 
			SET name = $1, category = $2, description = $3, price = $4
			WHERE product_id = $5
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.Name, testP.Category, testP.Description, testP.Price, testP.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(testP.ID, testP.Name, testP.Category, testP.Description, testP.Price, testP.CreatedAt),
		)

	result, err := repo.Update(context.Background(), testP)

	assert.NoError(t, err)
	assert.Equal(t, testP, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE products 
			SET name = $1, category = $2, description = $3, price = $4
			WHERE product_id = $5
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.Name, testP.Category, testP.Description, testP.Price, testP.ID).
		WillReturnError(errors.New("database unavailable"))

	result, err := repo.Update(context.Background(), testP)

	assert.Error(t, err)
	assert.EqualError(t, err, "update product: database unavailable")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenScanError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE products 
			SET name = $1, category = $2, description = $3, price = $4
			WHERE product_id = $5
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.Name, testP.Category, testP.Description, testP.Price, testP.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow("invalid-uuid", testP.Name, testP.Category, testP.Description, testP.Price, testP.CreatedAt),
		)

	result, err := repo.Update(context.Background(), testP)

	assert.Error(t, err)
	assert.EqualError(t, err, "update product: scanning value error for column 'product_id': Scan: invalid UUID length: 12")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdate_WhenRowsError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectQuery(`
			UPDATE products 
			SET name = $1, category = $2, description = $3, price = $4
			WHERE product_id = $5
			RETURNING product_id, name, category, description, price, created_at;`).
		WithArgs(testP.Name, testP.Category, testP.Description, testP.Price, testP.ID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				RowError(1, errors.New("read failed")),
		)

	result, err := repo.Update(context.Background(), testP)

	assert.Error(t, err)
	assert.EqualError(t, err, "update product: no rows in result set")
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_WhenNoDbIssues_Returns(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectExec(`
			DELETE FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), SecondUUID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_WhenDbError_ReturnsError(t *testing.T) {
	mock, repo := setup(t)
	mock.ExpectExec(`
			DELETE FROM products
			WHERE product_id = $1`).
		WithArgs(SecondUUID).
		WillReturnError(errors.New("database unavailable"))

	err := repo.Delete(context.Background(), SecondUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "delete product: database unavailable")
	assert.NoError(t, mock.ExpectationsWereMet())
}

var (
	FirstUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001")
	SecondUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002")
	ThirdUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d003")
	FourthUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d004")

	testP = product.Product{
		ID:          FirstUUID,
		Name:        "iPhone 15",
		Category:    "CLOTHES",
		Description: new("Updated description"),
		Price:       1500.0,
		CreatedAt:   time.Now(),
	}
)
