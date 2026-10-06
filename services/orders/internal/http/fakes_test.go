package http

import (
	"context"
	"testing"
	"time"

	"commerce-platform/services/orders/internal/grpc"
	"commerce-platform/services/orders/internal/order"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ---- GRPC fakes ----

type mockProductsClient struct {
	productIDs map[string]bool
}

func (m *mockProductsClient) GetProductByID(_ context.Context, id string) (*grpc.GetProductByIDResponse, error) {
	if m.productIDs[id] {
		return &grpc.GetProductByIDResponse{Id: id}, nil
	}
	return nil, status.Error(codes.NotFound, "product not found")
}

// ---- REPO fakes ----

type fakeOrderService struct {
	t            *testing.T
	getOrders    func(context.Context) ([]order.Order, error)
	getOrderByID func(context.Context, uuid.UUID) (order.Order, error)
	createOrder  func(context.Context, string, int) (order.Order, error)
	updateOrder  func(context.Context, uuid.UUID, string, int, order.OrderStatus) (order.Order, error)
	deleteOrder  func(context.Context, uuid.UUID) error

	foundIDs []uuid.UUID
	created  []order.Order
	updated  []order.Order
	deleted  []uuid.UUID

	pending map[string]int // track method pending
}

func newFakeOrderService(t *testing.T) *fakeOrderService {
	f := &fakeOrderService{t: t, pending: map[string]int{}}
	t.Cleanup(f.verify)
	return f
}

func (f *fakeOrderService) expectGetOrders(fn func(context.Context) ([]order.Order, error)) {
	f.getOrders = fn
	f.pending["GetOrders"]++
}

func (f *fakeOrderService) expectGetOrderByID(fn func(context.Context, uuid.UUID) (order.Order, error)) {
	f.getOrderByID = fn
	f.pending["GetOrderByID"]++
}

func (f *fakeOrderService) expectCreateOrder(
	fn func(ctx context.Context, productId string, quantity int) (order.Order, error)) {
	f.createOrder = fn
	f.pending["CreateOrder"]++
}

func (f *fakeOrderService) expectUpdateOrder(
	fn func(
		ctx context.Context,
		id uuid.UUID,
		productID string,
		quantity int,
		status order.OrderStatus,
	) (order.Order, error)) {
	f.updateOrder = fn
	f.pending["UpdateOrder"]++
}

func (f *fakeOrderService) expectDeleteOrder(fn func(context.Context, uuid.UUID) error) {
	f.deleteOrder = fn
	f.pending["DeleteOrder"]++
}

func (f *fakeOrderService) take(method string) {
	f.t.Helper()
	if f.pending[method] == 0 {
		f.t.Fatalf("unexpected call: %s", method)
	}
	f.pending[method]--
}

func (f *fakeOrderService) verify() {
	f.t.Helper()
	for method, n := range f.pending {
		if n != 0 {
			f.t.Errorf("repository %s: %d expected call(s) did not happen", method, n)
		}
	}
}

func (f *fakeOrderService) GetOrders(ctx context.Context) ([]order.Order, error) {
	f.take("GetOrders")
	return f.getOrders(ctx)
}

func (f *fakeOrderService) GetOrderByID(ctx context.Context, id uuid.UUID) (order.Order, error) {
	f.take("GetOrderByID")
	f.foundIDs = append(f.foundIDs, id)
	return f.getOrderByID(ctx, id)
}

func (f *fakeOrderService) CreateOrder(ctx context.Context, productID string, quantity int) (order.Order, error) {
	f.take("CreateOrder")
	o := order.Order{
		ID:        uuid.New(),
		ProductID: productID,
		Quantity:  quantity,
		Status:    order.CREATED,
		CreatedAt: time.Now(),
	}
	f.created = append(f.created, o)
	return f.createOrder(ctx, productID, quantity)
}

func (f *fakeOrderService) UpdateOrder(ctx context.Context, id uuid.UUID, productID string, quantity int, status order.OrderStatus) (order.Order, error) {
	f.take("UpdateOrder")
	o := order.Order{
		ID:        id,
		ProductID: productID,
		Quantity:  quantity,
		Status:    status,
	}
	f.updated = append(f.updated, o)
	return f.updateOrder(ctx, id, productID, quantity, status)
}

func (f *fakeOrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	f.take("DeleteOrder")
	f.deleted = append(f.deleted, id)
	return f.deleteOrder(ctx, id)
}
