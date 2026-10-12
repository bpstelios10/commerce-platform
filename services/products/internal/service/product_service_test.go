package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupProductServiceTest(t *testing.T) (*ProductServiceImpl, *fakeProductRepo) {
	t.Helper()
	repo := newFakeProductRepo(t)
	svc := NewProductServiceImpl(repo)

	return svc, repo
}

func TestGetProducts_WhenProductExists_ReturnsProducts(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, nil
		})

	p, err := svc.GetProducts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 4, len(p))
}

func TestGetProducts_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return []product.Product{}, errors.New("query products: database unavailable")
		})

	p, err := svc.GetProducts(context.Background())

	assert.Error(t, err)
	assert.EqualError(t, err, "get products: query products: database unavailable")
	assert.Nil(t, p)
}

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectGetByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			if id == FirstUUID {
				return product.Product{
					ID:          FirstUUID,
					Name:        "MacBook Pro",
					Category:    "ACCESSORY",
					Description: new("Apple laptop"),
					Price:       2500,
					CreatedAt:   time.Now(),
				}, nil
			}
			return product.Product{}, nil
		})

	p, err := svc.GetProductByID(context.Background(), FirstUUID)

	assert.NoError(t, err)
	assert.Equal(t, FirstUUID, p.ID)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, "ACCESSORY", p.Category)
	assert.Equal(t, new("Apple laptop"), p.Description)
	assert.Equal(t, 2500.0, p.Price)
	assert.WithinDuration(t, time.Now(), p.CreatedAt, time.Second)
	assert.Equal(t, FirstUUID, mock.searchedIDs[0])
}

func TestGetProductByID_WhenProductDoesNotExist_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	id, _ := uuid.NewV7()
	mock.expectGetByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			return product.Product{}, repository.ErrNotFound
		})

	p, err := svc.GetProductByID(context.Background(), id)

	var notFoundErr *ErrProductNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.ProductID)
	assert.EqualError(t, err, "Product with id ["+id.String()+"] was not found")
	assert.Empty(t, p)
	assert.Equal(t, mock.searchedIDs[0], id)
}

func TestGetProductByID_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	id, _ := uuid.NewV7()
	mock.expectGetByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			return product.Product{}, errors.New("find product by id: database unavailable")
		})

	p, err := svc.GetProductByID(context.Background(), id)

	assert.Error(t, err)
	assert.EqualError(t, err, "get product by id: find product by id: database unavailable")
	assert.Empty(t, p)
	assert.Equal(t, mock.searchedIDs[0], id)
}

func TestSearchProducts_WhenOnlyQueryProvided_FiltersByName(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, nil
		})

	products, err := svc.SearchProducts(context.Background(), "hoodie", nil, "")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, ThirdUUID, products[0].ID)
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_FiltersByBoth(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, nil
		})
	maxPrice := 200.0

	products, err := svc.SearchProducts(context.Background(), "necklace", &maxPrice, "")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, FourthUUID, products[0].ID)
}

func TestSearchProducts_WhenOnlyMaxPriceProvided_FiltersByPriceAndKeepsEqualBoundary(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, nil
		})
	maxPrice := 150.0

	products, err := svc.SearchProducts(context.Background(), "", &maxPrice, "")

	assert.NoError(t, err)
	assert.Len(t, products, 2)

	ids := []string{products[0].ID.String(), products[1].ID.String()}
	assert.Contains(t, ids, ThirdUUID.String())
	assert.Contains(t, ids, FourthUUID.String())
}

func TestSearchProducts_WhenOnlyCategoryProvided_FiltersByCategory(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, nil
		})

	products, err := svc.SearchProducts(context.Background(), "", nil, "accessory")

	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestSearchProducts_WhenAllCriteriaProvided_FiltersByCombinedCriteria(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, nil
		})
	maxPrice := 100.0

	products, err := svc.SearchProducts(context.Background(), "hoodie", &maxPrice, "clothes")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, ThirdUUID, products[0].ID)
}

func TestSearchProducts_WhenDbError_ReturnsError(t *testing.T) {
	svc, mock := setupProductServiceTest(t)
	mock.expectList(
		func(ctx context.Context) ([]product.Product, error) {
			return allProducts, errors.New("query products: database unavailable")
		})
	maxPrice := 100.0

	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	products, err := svc.SearchProducts(ctxWithError, "hoodie", &maxPrice, "clothes")

	assert.Error(t, err)
	assert.EqualError(t, err, "get products: query products: database unavailable")
	assert.Nil(t, products)
}

var (
	FirstUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d001")
	SecondUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d002")
	ThirdUUID  = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d003")
	FourthUUID = uuid.MustParse("f47ac10b-58cc-4372-a567-0e02b2c3d004")

	allProducts = []product.Product{
		{ID: FirstUUID, Name: "MacBook Pro", Category: "ACCESSORY", Description: new("Apple laptop"), Price: 2500, CreatedAt: time.Now()},
		{ID: SecondUUID, Name: "iPhone", Category: "ACCESSORY", Description: new("Apple smartphone"), Price: 1200, CreatedAt: time.Now()},
		{ID: ThirdUUID, Name: `hoodie Mykonos`, Category: "CLOTHES", Description: new("Comfortable hoodie"), Price: 80, CreatedAt: time.Now()},
		{ID: FourthUUID, Name: "Eye necklace", Category: "JEWELRY", Description: new("Stylish eye necklace"), Price: 150, CreatedAt: time.Now()},
	}
)
