package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"commerce-platform/services/orders/internal/order"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrVersionConflict = errors.New("order was updated by another transaction")
)

type Timeouts struct {
	Read, List, Write time.Duration
}

type PostgreOrderRepository struct {
	db       DB
	timeouts Timeouts
}

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func NewPostgreOrderRepository(db DB, t Timeouts) *PostgreOrderRepository {
	return &PostgreOrderRepository{db: db, timeouts: t}
}

func (repo *PostgreOrderRepository) List(ctx context.Context) ([]order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.timeouts.List)
	defer cancel()

	const query = `
		SELECT order_id, product_id, quantity, status, created_at, version
		FROM orders
	`

	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]order.Order, 0)

	for rows.Next() {
		var o order.Order

		if err := rows.Scan(
			&o.ID,
			&o.ProductID,
			&o.Quantity,
			&o.Status,
			&o.CreatedAt,
			&o.Version,
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

func (repo *PostgreOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.timeouts.Read)
	defer cancel()

	const query = `
		SELECT order_id, product_id, quantity, status, created_at, version
		FROM orders
		WHERE order_id = $1
	`
	var o order.Order

	err := repo.db.QueryRow(ctx, query, id).Scan(
		&o.ID,
		&o.ProductID,
		&o.Quantity,
		&o.Status,
		&o.CreatedAt,
		&o.Version,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return order.Order{}, fmt.Errorf("find order by id: %w", ErrNotFound)
	}

	if err != nil {
		return order.Order{}, fmt.Errorf("find order by id: %w", err)
	}

	return o, nil
}

func (repo *PostgreOrderRepository) Create(ctx context.Context, o order.Order) (order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.timeouts.Write)
	defer cancel()

	const query = `
		INSERT INTO orders (order_id, product_id, quantity, status)
		VALUES ($1, $2, $3, $4)
		RETURNING order_id, product_id, quantity, status, created_at, version;
	`

	row := repo.db.QueryRow(ctx, query, o.ID, o.ProductID, o.Quantity, o.Status)
	if err := row.Scan(
		&o.ID,
		&o.ProductID,
		&o.Quantity,
		&o.Status,
		&o.CreatedAt,
		&o.Version,
	); err != nil {
		return order.Order{}, fmt.Errorf("save order: %w", err)
	}

	return o, nil
}

func (repo *PostgreOrderRepository) Update(ctx context.Context, o order.Order) (order.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, repo.timeouts.Write)
	defer cancel()

	const query = `
		UPDATE orders
		SET product_id = $1, quantity = $2, status = $3, version = version + 1
		WHERE order_id = $4
		  AND version = $5
		RETURNING order_id, product_id, quantity, status, created_at, version;
	`

	row := repo.db.QueryRow(ctx, query, o.ProductID, o.Quantity, o.Status, o.ID, o.Version)
	err := row.Scan(
		&o.ID,
		&o.ProductID,
		&o.Quantity,
		&o.Status,
		&o.CreatedAt,
		&o.Version,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return order.Order{}, fmt.Errorf("update order: %w", ErrVersionConflict)
	}

	if err != nil {
		return order.Order{}, fmt.Errorf("update order: %w", err)
	}

	return o, nil
}

func (repo *PostgreOrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, repo.timeouts.Read)
	defer cancel()
	const query = `
		DELETE FROM orders
		WHERE order_id = $1
	`

	_, err := repo.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}

	return nil
}
