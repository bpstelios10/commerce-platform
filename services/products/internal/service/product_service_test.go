package service

import (
	"commerce-platform/services/products/internal/repository"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetProducts_WhenProductExists_ReturnsProducts(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProducts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 4, len(p))
}

func TestGetProducts_WhenDbError_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	p, err := svc.GetProducts(ctxWithError)

	assert.Error(t, err)
	assert.EqualError(t, err, "get products: unexpected error")
	assert.Nil(t, p)
}

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProductByID(context.Background(), repository.FirstUUID)

	assert.NoError(t, err)
	assert.Equal(t, repository.FirstUUID, p.ID)
	assert.Equal(t, "MacBook Pro", p.Name)
	assert.Equal(t, "ACCESSORY", p.Category)
	assert.Equal(t, 2500.0, p.Price)
	assert.Equal(t, 10, p.Stock)
}

func TestGetProductByID_WhenProductDoesNotExist_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	id, _ := uuid.NewV7()

	p, err := svc.GetProductByID(context.Background(), id)

	var notFoundErr *ErrProductNotFound
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.ProductID)
	assert.Equal(t, "Product with id ["+id.String()+"] was not found", notFoundErr.Error())
	assert.Empty(t, p)
}

func TestGetProductByID_WhenDbError_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProductByID(context.Background(), repository.ErrornousUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "get product by id: unexpected error")
	assert.Empty(t, p)
}

func TestSearchProducts_WhenOnlyQueryProvided_FiltersByName(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	products, err := svc.SearchProducts(context.Background(), "hoodie", nil, "")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, repository.ThirdUUID, products[0].ID)
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_FiltersByBoth(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 200.0

	products, err := svc.SearchProducts(context.Background(), "necklace", &maxPrice, "")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, repository.FourthUUID, products[0].ID)
}

func TestSearchProducts_WhenOnlyMaxPriceProvided_FiltersByPriceAndKeepsEqualBoundary(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 150.0

	products, err := svc.SearchProducts(context.Background(), "", &maxPrice, "")

	assert.NoError(t, err)
	assert.Len(t, products, 2)

	ids := []string{products[0].ID.String(), products[1].ID.String()}
	assert.Contains(t, ids, repository.ThirdUUID.String())
	assert.Contains(t, ids, repository.FourthUUID.String())
}

func TestSearchProducts_WhenOnlyCategoryProvided_FiltersByCategory(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	products, err := svc.SearchProducts(context.Background(), "", nil, "accessory")

	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestSearchProducts_WhenAllCriteriaProvided_FiltersByCombinedCriteria(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 100.0

	products, err := svc.SearchProducts(context.Background(), "hoodie", &maxPrice, "clothes")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, repository.ThirdUUID, products[0].ID)
}

func TestSearchProducts_WhenDbError_ReturnsError(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 100.0

	ctx := context.Background()
	ctxWithError := context.WithValue(ctx, "errorEnabler", "unexpected error")

	products, err := svc.SearchProducts(ctxWithError, "hoodie", &maxPrice, "clothes")

	assert.Error(t, err)
	assert.EqualError(t, err, "get products: unexpected error")
	assert.Nil(t, products)
}
