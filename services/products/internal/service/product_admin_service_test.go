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

func setup(t *testing.T) (*AdminService, *fakeProductRepo, *fakeProductCategoryRepo) {
	t.Helper()

	repo := newFakeProductRepo(t)
	productService := NewProductServiceImpl(repo)

	categoryRepo := newFakeProductCategoryRepo(t)
	categoryService := NewProductCategoryService(categoryRepo)
	svc := NewAdminService(productService, categoryService, repo)

	return svc, repo, categoryRepo
}

func TestCreateProduct_WhenProductNotExists(t *testing.T) {
	svc, productRepo, categRepo := setup(t)
	p := product.Product{
		Name:        "MacBook Pro M4",
		Category:    "ACCESSORY",
		Description: new("some-description"),
		Price:       2501.0,
	}
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return true, nil
		})
	productRepo.expectSave(
		func(ctx context.Context, p product.Product) (product.Product, error) {
			p.CreatedAt = time.Now()
			return p, nil
		})

	p, err := svc.CreateProduct(context.Background(), p.Name, p.Category, p.Description, p.Price)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, p.Category, categRepo.searchedCategories[0])
	assert.Equal(t, p.Name, productRepo.saved[0].Name)
	assert.NotEmpty(t, productRepo.saved[0].ID) // the id should be created in the service
}

func TestCreateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, _, categRepo := setup(t)
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return false, nil
		})

	p, err := svc.CreateProduct(context.Background(), "MacBook Pro M4", "UNKNOWN", new("some-description"), 2501.0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.EqualError(t, err, "validate product category UNKNOWN: invalid category")
	assert.Empty(t, p)
	assert.Equal(t, "UNKNOWN", categRepo.searchedCategories[0])
}

type failingReader struct {
	err error
}

func (r failingReader) Read([]byte) (int, error) {
	return 0, r.err
}

func TestCreateProduct_WhenUUIDGenerationFails_ReturnsError(t *testing.T) {
	entropyErr := errors.New("entropy unavailable")
	uuid.SetRand(failingReader{err: entropyErr})
	t.Cleanup(func() { uuid.SetRand(nil) }) // resets to crypto/rand.Reader
	svc, _, categRepo := setup(t)
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return true, nil
		})

	_, err := svc.CreateProduct(context.Background(), "name", "CATEGORY", nil, 1)

	assert.ErrorIs(t, err, entropyErr)
	assert.EqualError(t, err, "create product: entropy unavailable")
	assert.Equal(t, "CATEGORY", categRepo.searchedCategories[0])
}

func TestCreateProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, productRepo, categRepo := setup(t)
	existing := product.Product{
		Name:        "MacBook Pro M4",
		Category:    "ACCESSORY",
		Description: new("some-description"),
		Price:       2501.0,
	}
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return true, nil
		})
	productRepo.expectSave(
		func(ctx context.Context, p product.Product) (product.Product, error) {
			return product.Product{}, errors.New("save product: database unavailable")
		})

	p, err := svc.CreateProduct(
		context.Background(),
		existing.Name, existing.Category, existing.Description, existing.Price)

	assert.Error(t, err)
	assert.EqualError(t, err, "create product: save product: database unavailable")
	assert.Empty(t, p)
	assert.Equal(t, "ACCESSORY", categRepo.searchedCategories[0])
	assert.Equal(t, existing.Name, productRepo.saved[0].Name)
}

func TestUpdateProduct_WhenProductNotExists_Returns404(t *testing.T) {
	svc, productRepo, _ := setup(t)
	id, _ := uuid.NewV7()
	productRepo.expectFindByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			return product.Product{}, repository.ErrNotFound
		})

	updated, err := svc.UpdateProduct(context.Background(), id, "whatever", "ACCESSORY", new(""), 1201.0)

	var notFoundErr *ErrProductNotFound
	assert.Error(t, err)
	assert.ErrorAs(t, err, &notFoundErr)
	assert.Equal(t, id, notFoundErr.ProductID)
	assert.EqualError(t, err, "updating product: Product with id ["+id.String()+"] was not found")
	assert.Empty(t, updated)
	assert.Equal(t, id, productRepo.searchedIDs[0])
}

func TestUpdateProduct_WhenProductExists_UpdatesProduct(t *testing.T) {
	svc, productRepo, categRepo := setup(t)
	createdAt := time.Now()
	p := &product.Product{
		ID:          SecondUUID,
		Name:        "iPhone 15",
		Category:    "CLOTHES",
		Description: new("Updated description"),
		Price:       1500.0,
		CreatedAt:   createdAt,
	}
	productRepo.expectFindByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			p.CreatedAt = createdAt
			return product.Product{
				ID:          SecondUUID,
				Name:        "iPhone",
				Category:    "ACCESSORY",
				Description: new("Apple smartphone"),
				Price:       1200,
				CreatedAt:   createdAt}, nil
		})
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return true, nil
		})
	productRepo.expectUpdate(
		func(ctx context.Context, p product.Product) (product.Product, error) {
			return p, nil
		})

	updated, err := svc.UpdateProduct(context.Background(), p.ID, p.Name, p.Category, p.Description, p.Price)

	assert.NoError(t, err)
	assert.Equal(t, p.ID, updated.ID)
	assert.Equal(t, p.Name, updated.Name)
	assert.Equal(t, p.Category, updated.Category)
	assert.Equal(t, p.Description, updated.Description)
	assert.Equal(t, p.Price, updated.Price)
	assert.True(t, updated.CreatedAt.Equal(createdAt))
	assert.Equal(t, p.Category, categRepo.searchedCategories[0])
	assert.Equal(t, p.ID, productRepo.searchedIDs[0])
	assert.Equal(t, p.Name, productRepo.updated[0].Name)
}

func TestUpdateProduct_WhenCategoryInvalid_ReturnsInvalidCategory(t *testing.T) {
	svc, productRepo, categRepo := setup(t)
	productRepo.expectFindByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			return product.Product{
				ID:          SecondUUID,
				Name:        "iPhone",
				Category:    "ACCESSORY",
				Description: new("Apple smartphone"),
				Price:       1200,
				CreatedAt:   time.Now()}, nil
		})
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return false, nil
		})

	updated, err := svc.UpdateProduct(context.Background(), SecondUUID, "iPhone 7", "UNKNOWN", new(""), 1201.0)

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidCategory)
	assert.EqualError(t, err, "updating product: validate product category UNKNOWN: invalid category")
	assert.Empty(t, updated)
	assert.Equal(t, SecondUUID, productRepo.searchedIDs[0])
	assert.Equal(t, "UNKNOWN", categRepo.searchedCategories[0])
}

func TestUpdateProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, productRepo, categRepo := setup(t)
	productRepo.expectFindByID(
		func(ctx context.Context, id uuid.UUID) (product.Product, error) {
			return product.Product{
				ID:          SecondUUID,
				Name:        "iPhone",
				Category:    "ACCESSORY",
				Description: new("Apple smartphone"),
				Price:       1200,
				CreatedAt:   time.Now()}, nil
		})
	categRepo.expectExists(
		func(ctx context.Context, s string) (bool, error) {
			return true, nil
		})
	productRepo.expectUpdate(
		func(ctx context.Context, p product.Product) (product.Product, error) {
			return p, errors.New("update product: database unavailable")
		})

	updated, err := svc.UpdateProduct(context.Background(), SecondUUID, "iPhone 7", "CLOTHES", new(""), 1201.0)

	assert.Error(t, err)
	assert.EqualError(t, err, "updating product: update product: database unavailable")
	assert.Empty(t, updated)
	assert.Equal(t, "CLOTHES", categRepo.searchedCategories[0])
	assert.Equal(t, SecondUUID, productRepo.searchedIDs[0])
	assert.Equal(t, "iPhone 7", productRepo.updated[0].Name)
}

func TestDeleteProduct_WhenProductNotExists_DoesNotFail(t *testing.T) {
	svc, productRepo, _ := setup(t)
	id, _ := uuid.NewV7()
	productRepo.expectDelete(
		func(ctx context.Context, u uuid.UUID) error {
			return nil
		})

	err := svc.DeleteProduct(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, id, productRepo.deleted[0])
}

func TestDeleteProduct_WhenProductExists_DeletesProduct(t *testing.T) {
	svc, productRepo, _ := setup(t)
	productRepo.expectDelete(
		func(ctx context.Context, u uuid.UUID) error {
			return nil
		})

	err := svc.DeleteProduct(context.Background(), SecondUUID)

	assert.NoError(t, err)
	assert.Equal(t, SecondUUID, productRepo.deleted[0])
}

func TestDeleteProduct_WhenDbError_ReturnsError(t *testing.T) {
	svc, productRepo, _ := setup(t)
	productRepo.expectDelete(
		func(ctx context.Context, u uuid.UUID) error {
			return errors.New("delete product: database unavailable")
		})

	err := svc.DeleteProduct(context.Background(), SecondUUID)

	assert.Error(t, err)
	assert.EqualError(t, err, "deleting product: delete product: database unavailable")
	assert.Equal(t, SecondUUID, productRepo.deleted[0])
}
