package repository

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/services/orders/internal/order"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("not found")

type PostgreOrderRepository struct {
	db DB
}

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewPostgreOrderRepository(db DB) *PostgreOrderRepository {
	return &PostgreOrderRepository{db: db}
}

func (repo *PostgreOrderRepository) FindAll(ctx context.Context) []order.Order {
	var orders []order.Order

	return orders
}

func (repo *PostgreOrderRepository) FindByID(ctx context.Context, id string) (order.Order, error) {
	const query = `
		SELECT id, customer_id, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var o order.Order

	err := repo.db.QueryRow(ctx, query, id).Scan(
		&o.ID,
		&o.ProductID,
		&o.Quantity,
		&o.Status,
		// &o.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return order.Order{}, ErrNotFound
	}

	if err != nil {
		return order.Order{}, fmt.Errorf("find order by id: %w", err)
	}

	return o, nil
}

func (repo *PostgreOrderRepository) Save(ctx context.Context, o order.Order) {
}

func (repo *PostgreOrderRepository) Update(ctx context.Context, o order.Order) {
}

func (repo *PostgreOrderRepository) Delete(ctx context.Context, id uuid.UUID) {
}
