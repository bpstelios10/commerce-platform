package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetProducts_WhenProductExists_ReturnsProducts(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p := svc.GetProducts()

	assert.Equal(t, 4, len(p))
}

func TestGetProductByID_WhenProductExists_ReturnsProduct(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	p, err := svc.GetProductByID(repository.FirstUUID)

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

	p, err := svc.GetProductByID(id)

	assert.ErrorIs(t, err, ErrProductNotFound)
	assert.Equal(t, product.Product{}, p)
}

func TestSearchProducts_WhenOnlyQueryProvided_FiltersByName(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	products := svc.SearchProducts("hoodie", nil, "")

	assert.Len(t, products, 1)
	assert.Equal(t, repository.ThirdUUID, products[0].ID)
}

func TestSearchProducts_WhenQueryAndMaxPriceProvided_FiltersByBoth(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 200.0

	products := svc.SearchProducts("necklace", &maxPrice, "")

	assert.Len(t, products, 1)
	assert.Equal(t, repository.FourthUUID, products[0].ID)
}

func TestSearchProducts_WhenOnlyMaxPriceProvided_FiltersByPriceAndKeepsEqualBoundary(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 150.0

	products := svc.SearchProducts("", &maxPrice, "")

	assert.Len(t, products, 2)

	ids := []string{products[0].ID.String(), products[1].ID.String()}
	assert.Contains(t, ids, repository.ThirdUUID.String())
	assert.Contains(t, ids, repository.FourthUUID.String())
}

func TestSearchProducts_WhenOnlyCategoryProvided_FiltersByCategory(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)

	products := svc.SearchProducts("", nil, "accessory")

	assert.Len(t, products, 2)
}

func TestSearchProducts_WhenAllCriteriaProvided_FiltersByCombinedCriteria(t *testing.T) {
	repo := repository.NewInMemoryProductRepository()
	svc := NewProductService(repo)
	maxPrice := 100.0

	products := svc.SearchProducts("hoodie", &maxPrice, "clothes")

	assert.Len(t, products, 1)
	assert.Equal(t, repository.ThirdUUID, products[0].ID)
}
