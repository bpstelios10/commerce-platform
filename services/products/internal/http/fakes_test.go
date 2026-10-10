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

	searchedIDs []uuid.UUID

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
