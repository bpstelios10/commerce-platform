package service

import (
	"commerce-platform/services/products/internal/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func setupProductServiceTest(t *testing.T) (*ProductService, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	assert.NoError(t, err)

	t.Cleanup(func() {
		mock.Close()
	})
	repo := repository.NewPostgreProductRepository(mock)
	svc := NewProductService(repo)

	return svc, mock
}

func TestGetProducts_WhenProductExists_ReturnsProducts(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
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

	p, err := svc.GetProducts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 4, len(p))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProducts_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnError(errors.New("database unavailable"))

	p, err := svc.GetProducts(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "get products: query products: database unavailable")
	assert.Nil(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(FirstUUID).
		WillReturnRows(
			pgxmock.NewRows([]string{
				"product_id", "name", "category", "description", "price", "created_at",
			}).
				AddRow(FirstUUID, "MacBook Pro", "ACCESSORY", new("Apple laptop"), 2500, time.Now()),
		)

	p, err := svc.GetProductByID(context.Background(), FirstUUID)

	assert.NoError(t, err)
	assert.Equal(t, FirstUUID, p.ID)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, "ACCESSORY", p.Category)
	assert.Equal(t, new("Apple laptop"), p.Description)
	assert.Equal(t, 2500.0, p.Price)
	assert.WithinDuration(t, time.Now(), p.CreatedAt, time.Second)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductByID_WhenProductDoesNotExist_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnRows(pgxmock.NewRows([]string{
			"product_id", "name", "category", "description", "price", "created_at",
		}))

	p, err := svc.GetProductByID(context.Background(), id)

	var notFoundErr *ErrProductNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.ProductID)
	assert.EqualError(t, err, "Product with id ["+id.String()+"] was not found")
	assert.Empty(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProductByID_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	id, _ := uuid.NewV7()
	mock.ExpectQuery(`
			SELECT product_id, name, category, description, price, created_at
			FROM products
			WHERE product_id = $1`).
		WithArgs(id).
		WillReturnError(errors.New("database unavailable"))

	p, err := svc.GetProductByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "get product by id: find product by id: database unavailable")
	assert.Empty(t, p)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenOnlyQueryProvided_FiltersByName(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
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

	products, err := svc.SearchProducts(context.Background(), "hoodie", nil, "")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, ThirdUUID, products[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_FiltersByBoth(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
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
	maxPrice := 200.0

	products, err := svc.SearchProducts(context.Background(), "necklace", &maxPrice, "")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, FourthUUID, products[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenOnlyMaxPriceProvided_FiltersByPriceAndKeepsEqualBoundary(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
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
	maxPrice := 150.0

	products, err := svc.SearchProducts(context.Background(), "", &maxPrice, "")

	assert.NoError(t, err)
	assert.Len(t, products, 2)

	ids := []string{products[0].ID.String(), products[1].ID.String()}
	assert.Contains(t, ids, ThirdUUID.String())
	assert.Contains(t, ids, FourthUUID.String())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenOnlyCategoryProvided_FiltersByCategory(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
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

	products, err := svc.SearchProducts(context.Background(), "", nil, "accessory")

	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenAllCriteriaProvided_FiltersByCombinedCriteria(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
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
	maxPrice := 100.0

	products, err := svc.SearchProducts(context.Background(), "hoodie", &maxPrice, "clothes")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, ThirdUUID, products[0].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSearchProducts_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.ExpectQuery(`SELECT product_id, name, category, description, price, created_at FROM products`).
		WillReturnError(errors.New("database unavailable"))
	maxPrice := 100.0

	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	products, err := svc.SearchProducts(ctxWithError, "hoodie", &maxPrice, "clothes")

	assert.Error(t, err)
	assert.EqualError(t, err, "get products: query products: database unavailable")
	assert.Nil(t, products)
	assert.NoError(t, mock.ExpectationsWereMet())
}

var (
	FirstUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001")
	SecondUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002")
	ThirdUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d003")
	FourthUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d004")
)
