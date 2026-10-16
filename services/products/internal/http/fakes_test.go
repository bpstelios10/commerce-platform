package http

import (
	"commerce-platform/services/products/internal/product"
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeProductService struct {
	t              *testing.T
	getProducts    func(context.Context) ([]product.Product, error)
	getProductByID func(context.Context, uuid.UUID) (product.Product, error)
	searchProducts func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error)
	createProduct  func(ctx context.Context, name string, category string, description *string, price float64) (product.Product, error)
	updateProduct  func(ctx context.Context, id uuid.UUID, name string, category string, description *string, price float64, version int) (product.Product, error)
	deleteProduct  func(context.Context, uuid.UUID) error

	// capture the input params, so that i can validate they are the expected ones
	searchedIDs []uuid.UUID
	created     []product.Product
	updated     []product.Product
	deleted     []uuid.UUID

	pending map[string]int // track method pending
}

func newFakeProductService(t *testing.T) *fakeProductService {
	f := &fakeProductService{t: t, pending: map[string]int{}}
	t.Cleanup(f.verify)
	return f
}

func (f *fakeProductService) expectGetProducts(fn func(context.Context) ([]product.Product, error)) {
	f.getProducts = fn
	f.pending["GetProducts"]++
}

func (f *fakeProductService) expectGetProductByID(fn func(context.Context, uuid.UUID) (product.Product, error)) {
	f.getProductByID = fn
	f.pending["GetProductByID"]++
}

func (f *fakeProductService) expectSearchProducts(
	fn func(ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error)) {
	f.searchProducts = fn
	f.pending["SearchProducts"]++
}

func (f *fakeProductService) expectCreateProduct(
	fn func(ctx context.Context, name string, category string, description *string, price float64) (product.Product, error)) {
	f.createProduct = fn
	f.pending["CreateProduct"]++
}

func (f *fakeProductService) expectUpdateProduct(
	fn func(ctx context.Context, id uuid.UUID, name string, category string, description *string, price float64, version int) (product.Product, error)) {
	f.updateProduct = fn
	f.pending["UpdateProduct"]++
}

func (f *fakeProductService) expectDeleteProduct(fn func(context.Context, uuid.UUID) error) {
	f.deleteProduct = fn
	f.pending["DeleteProduct"]++
}

func (f *fakeProductService) take(method string) {
	f.t.Helper()
	if f.pending[method] == 0 {
		f.t.Fatalf("unexpected call: %s", method)
	}
	f.pending[method]--
}

func (f *fakeProductService) verify() {
	f.t.Helper()
	for method, n := range f.pending {
		if n != 0 {
			f.t.Errorf("repository %s: %d expected call(s) did not happen", method, n)
		}
	}
}

func (f *fakeProductService) GetProducts(ctx context.Context) ([]product.Product, error) {
	f.take("GetProducts")
	return f.getProducts(ctx)
}

func (f *fakeProductService) GetProductByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	f.take("GetProductByID")
	f.searchedIDs = append(f.searchedIDs, id)
	return f.getProductByID(ctx, id)
}

func (f *fakeProductService) SearchProducts(
	ctx context.Context, query string, maxPrice *float64, category string) ([]product.Product, error) {
	f.take("SearchProducts")
	return f.searchProducts(ctx, query, maxPrice, category)
}

func (f *fakeProductService) CreateProduct(
	ctx context.Context,
	name, category string,
	description *string,
	price float64,
) (product.Product, error) {
	f.take("CreateProduct")
	f.created = append(f.created, product.Product{
		Name:        name,
		Category:    category,
		Description: description,
		Price:       price})
	return f.createProduct(ctx, name, category, description, price)
}

func (f *fakeProductService) UpdateProduct(
	ctx context.Context,
	id uuid.UUID,
	name, category string,
	description *string,
	price float64,
	version int,
) (product.Product, error) {
	f.take("UpdateProduct")
	f.updated = append(f.updated, product.Product{
		ID:          id,
		Name:        name,
		Category:    category,
		Description: description,
		Price:       price,
		Version:     version})

	return f.updateProduct(ctx, id, name, category, description, price, version+1)
}

func (f *fakeProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	f.take("DeleteProduct")
	f.deleted = append(f.deleted, id)

	return f.deleteProduct(ctx, id)
}

// ---- Product Category Fake ---- //

type fakeProductCategoryService struct {
	t                    *testing.T
	validate             func(context.Context, string) (string, error)
	getProductCategories func(context.Context) ([]string, error)

	// capture the input params, so that i can validate they are the expected ones
	searchedCategories []string

	pending map[string]int // track method pending
}

func newFakeProductCategoryService(t *testing.T) *fakeProductCategoryService {
	f := &fakeProductCategoryService{t: t, pending: map[string]int{}}
	t.Cleanup(f.verify)
	return f
}

func (f *fakeProductCategoryService) expectValidate(fn func(context.Context, string) (string, error)) {
	f.validate = fn
	f.pending["Validate"]++
}

func (f *fakeProductCategoryService) expectGetProductCategories(fn func(context.Context) ([]string, error)) {
	f.getProductCategories = fn
	f.pending["GetProductCategories"]++
}

func (f *fakeProductCategoryService) take(method string) {
	f.t.Helper()
	if f.pending[method] == 0 {
		f.t.Fatalf("unexpected call: %s", method)
	}
	f.pending[method]--
}

func (f *fakeProductCategoryService) verify() {
	f.t.Helper()
	for method, n := range f.pending {
		if n != 0 {
			f.t.Errorf("repository %s: %d expected call(s) did not happen", method, n)
		}
	}
}

func (f *fakeProductCategoryService) Validate(ctx context.Context, category string) (string, error) {
	f.take("Validate")
	f.searchedCategories = append(f.searchedCategories, category)
	return f.validate(ctx, category)
}

func (f *fakeProductCategoryService) GetProductCategories(ctx context.Context) ([]string, error) {
	f.take("GetProductCategories")
	return f.getProductCategories(ctx)
}
