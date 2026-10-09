package service

import (
	"commerce-platform/services/products/internal/product"
	"context"
	"testing"

	"github.com/google/uuid"
)

// ---- Product Categories ---- //

type fakeProductCategoryRepo struct {
	t      *testing.T
	exists func(context.Context, string) (bool, error)
	getAll func(context.Context) ([]string, error)

	foundCategories []string

	pending map[string]int // track method pending
}

func newFakeProductCategoryRepo(t *testing.T) *fakeProductCategoryRepo {
	f := &fakeProductCategoryRepo{t: t, pending: map[string]int{}}
	t.Cleanup(f.verify)
	return f
}

func (f *fakeProductCategoryRepo) expectExists(fn func(context.Context, string) (bool, error)) {
	f.exists = fn
	f.pending["exists"]++
}

func (f *fakeProductCategoryRepo) expectGetAll(fn func(context.Context) ([]string, error)) {
	f.getAll = fn
	f.pending["GetAll"]++
}

func (f *fakeProductCategoryRepo) take(method string) {
	f.t.Helper()
	if f.pending[method] == 0 {
		f.t.Fatalf("unexpected call: %s", method)
	}
	f.pending[method]--
}

func (f *fakeProductCategoryRepo) verify() {
	f.t.Helper()
	for method, n := range f.pending {
		if n != 0 {
			f.t.Errorf("repository %s: %d expected call(s) did not happen", method, n)
		}
	}
}

func (f *fakeProductCategoryRepo) Exists(ctx context.Context, category string) (bool, error) {
	f.take("exists")
	f.foundCategories = append(f.foundCategories, category)
	return f.exists(ctx, category)
}

func (f *fakeProductCategoryRepo) GetAll(ctx context.Context) ([]string, error) {
	f.take("GetAll")
	return f.getAll(ctx)
}

// ---- Products ---- //

type fakeProductRepo struct {
	t        *testing.T
	findAll  func(context.Context) ([]product.Product, error)
	findByID func(context.Context, uuid.UUID) (product.Product, error)

	foundIDs []uuid.UUID

	pending map[string]int // track method pending
}

func newFakeProductRepo(t *testing.T) *fakeProductRepo {
	f := &fakeProductRepo{t: t, pending: map[string]int{}}
	t.Cleanup(f.verify)
	return f
}

func (f *fakeProductRepo) expectFindAll(fn func(context.Context) ([]product.Product, error)) {
	f.findAll = fn
	f.pending["FindAll"]++
}

func (f *fakeProductRepo) expectFindByID(fn func(context.Context, uuid.UUID) (product.Product, error)) {
	f.findByID = fn
	f.pending["FindByID"]++
}

func (f *fakeProductRepo) take(method string) {
	f.t.Helper()
	if f.pending[method] == 0 {
		f.t.Fatalf("unexpected call: %s", method)
	}
	f.pending[method]--
}

func (f *fakeProductRepo) verify() {
	f.t.Helper()
	for method, n := range f.pending {
		if n != 0 {
			f.t.Errorf("repository %s: %d expected call(s) did not happen", method, n)
		}
	}
}

func (f *fakeProductRepo) FindAll(ctx context.Context) ([]product.Product, error) {
	f.take("FindAll")
	return f.findAll(ctx)
}

func (f *fakeProductRepo) FindByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	f.take("FindByID")
	p, err := f.findByID(ctx, id)
	if p != (product.Product{}) {
		f.foundIDs = append(f.foundIDs, id)
	}
	return p, err
}
