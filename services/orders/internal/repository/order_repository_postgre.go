package repository

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/services/orders/internal/order"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")

type PostgreOrderRepository struct {
	db DB
}

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func NewPostgreOrderRepository(db DB) *PostgreOrderRepository {
	return &PostgreOrderRepository{db: db}
}

func (repo *PostgreOrderRepository) FindAll(ctx context.Context) ([]order.Order, error) {
	const query = `
		SELECT id, product_id, quantity, status, created_at
		FROM orders
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	var orders []order.Order

	for rows.Next() {
		var o order.Order

		if err := rows.Scan(
			&o.ID,
			&o.ProductID,
			&o.Quantity,
			&o.Status,
			// &o.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}

		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}

	return orders, nil
}

func (repo *PostgreOrderRepository) FindByID(ctx context.Context, id string) (order.Order, error) {
	const query = `
		SELECT id, product_id, quantity, status, created_at
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

func (repo *PostgreOrderRepository) Save(ctx context.Context, o order.Order) error {
	const query = `
		INSERT INTO orders (id, product_id, quantity, status)
		VALUES ($1, $2, $3, $4)
	`

	_, err := repo.db.Exec(
		ctx,
		query,
		o.ID,
		o.ProductID,
		o.Quantity,
		o.Status,
	)

	return err
}

func (repo *PostgreOrderRepository) Update(ctx context.Context, o order.Order) error {
	const query = `
		UPDATE orders
		SET product_id = $1, quantity = $2, status = $3
		WHERE id = $4
	`

	_, err := repo.db.Exec(
		ctx,
		query,
		o.ProductID,
		o.Quantity,
		o.Status,
		o.ID,
	)

	return err
}

func (repo *PostgreOrderRepository) Delete(ctx context.Context, id uuid.UUID) {
}
