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

	// capture the input params, so that i can validate they are the expected ones
	searchedCategories []string

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
	f.searchedCategories = append(f.searchedCategories, category)
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
	save     func(context.Context, product.Product) (product.Product, error)
	update   func(context.Context, product.Product) (product.Product, error)
	delete   func(context.Context, uuid.UUID) error

	// capture the input params, so that i can validate they are the expected ones
	searchedIDs []uuid.UUID
	saved       []product.Product
	updated     []product.Product
	deleted     []uuid.UUID

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

func (f *fakeProductRepo) expectSave(fn func(context.Context, product.Product) (product.Product, error)) {
	f.save = fn
	f.pending["Save"]++
}

func (f *fakeProductRepo) expectUpdate(fn func(context.Context, product.Product) (product.Product, error)) {
	f.update = fn
	f.pending["Update"]++
}

func (f *fakeProductRepo) expectDelete(fn func(context.Context, uuid.UUID) error) {
	f.delete = fn
	f.pending["Delete"]++
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
	f.searchedIDs = append(f.searchedIDs, id)
	return f.findByID(ctx, id)
}

func (f *fakeProductRepo) Save(ctx context.Context, o product.Product) (product.Product, error) {
	f.take("Save")
	f.saved = append(f.saved, o)
	return f.save(ctx, o)
}

func (f *fakeProductRepo) Update(ctx context.Context, o product.Product) (product.Product, error) {
	f.take("Update")
	f.updated = append(f.updated, o)
	return f.update(ctx, o)
}

func (f *fakeProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	f.take("Delete")
	f.deleted = append(f.deleted, id)
	return f.delete(ctx, id)
}
