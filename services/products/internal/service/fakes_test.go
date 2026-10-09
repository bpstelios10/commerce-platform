package service

import (
	"context"
	"testing"
)

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
