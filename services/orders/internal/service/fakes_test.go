package service

import (
	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
)

type fakeOrderRepo struct {
	t        *testing.T
	findAll  func(context.Context) ([]order.Order, error)
	findByID func(context.Context, uuid.UUID) (order.Order, error)
	save     func(context.Context, order.Order) (order.Order, error)
	update   func(context.Context, order.Order) (order.Order, error)
	delete   func(context.Context, uuid.UUID) error

	foundIDs []uuid.UUID
	saved    []order.Order
	updated  []order.Order
	deleted  []uuid.UUID

	pending map[string]int // track method pending
}

func newFakeOrderRepo(t *testing.T) *fakeOrderRepo {
	f := &fakeOrderRepo{t: t, pending: map[string]int{}}
	t.Cleanup(f.verify)
	return f
}

func (f *fakeOrderRepo) expectFindAll(fn func(context.Context) ([]order.Order, error)) {
	f.findAll = fn
	f.pending["FindAll"]++
}

func (f *fakeOrderRepo) expectFindByID(fn func(context.Context, uuid.UUID) (order.Order, error)) {
	f.findByID = fn
	f.pending["FindByID"]++
}

func (f *fakeOrderRepo) expectSave(fn func(context.Context, order.Order) (order.Order, error)) {
	f.save = fn
	f.pending["Save"]++
}

func (f *fakeOrderRepo) expectUpdate(fn func(context.Context, order.Order) (order.Order, error)) {
	f.update = fn
	f.pending["Update"]++
}

func (f *fakeOrderRepo) expectDelete(fn func(context.Context, uuid.UUID) error) {
	f.delete = fn
	f.pending["Delete"]++
}

func (f *fakeOrderRepo) take(method string) {
	f.t.Helper()
	if f.pending[method] == 0 {
		f.t.Fatalf("unexpected call: %s", method)
	}
	f.pending[method]--
}

func (f *fakeOrderRepo) verify() {
	f.t.Helper()
	for method, n := range f.pending {
		if n != 0 {
			f.t.Errorf("repository %s: %d expected call(s) did not happen", method, n)
		}
	}
}

func (f *fakeOrderRepo) FindAll(ctx context.Context) ([]order.Order, error) {
	f.take("FindAll")
	return f.findAll(ctx)
}

func (f *fakeOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (order.Order, error) {
	f.take("FindByID")
	f.foundIDs = append(f.foundIDs, id)
	return f.findByID(ctx, id)
}

func (f *fakeOrderRepo) Save(ctx context.Context, o order.Order) (order.Order, error) {
	f.take("Save")
	f.saved = append(f.saved, o)
	return f.save(ctx, o)
}

func (f *fakeOrderRepo) Update(ctx context.Context, o order.Order) (order.Order, error) {
	f.take("Update")
	f.updated = append(f.updated, o)
	return f.update(ctx, o)
}

func (f *fakeOrderRepo) Delete(ctx context.Context, id uuid.UUID) error {
	f.take("Delete")
	f.deleted = append(f.deleted, id)
	return f.delete(ctx, id)
}

// ---- GRPC fakes ----

type fakeProductsClient struct {
	productIDs map[string]bool // product IDs that "exist"
}

func (f *fakeProductsClient) GetProductByID(_ context.Context, id string) (*grpc.GetProductByIDResponse, error) {
	if f.productIDs[id] {
		return &grpc.GetProductByIDResponse{Id: id}, nil
	} else if id == "error" {
		return nil, status.Error(codes.Internal, "unexpected error")
	}
	return nil, status.Error(codes.NotFound, "product not found")
}

func echoSave(_ context.Context, o order.Order) (order.Order, error) { return o, nil }
