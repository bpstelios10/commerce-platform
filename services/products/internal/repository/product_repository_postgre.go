package repository

import (
	"context"
	"errors"
	"fmt"

	"commerce-platform/services/products/internal/product"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")

type PostgreProductRepository struct {
	db ProductCategoryDB
}

type ProductDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func NewPostgreProductRepository(db ProductCategoryDB) *PostgreProductRepository {
	return &PostgreProductRepository{db: db}
}

func (r *PostgreProductRepository) FindAll(ctx context.Context) ([]product.Product, error) {
	// SELECT product_id, name, category, description, price, created_at
	const query = `
		SELECT product_id, name, category, price
		FROM products
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}

	defer rows.Close()

	var products []product.Product

	for rows.Next() {
		var p product.Product

		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Category,
			// &p.Description,
			&p.Price,
			// &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}

	return products, nil
}

func (r *PostgreProductRepository) FindByID(ctx context.Context, id uuid.UUID) (product.Product, error) {
	// SELECT product_id, name, category, description, price, created_at
	const query = `
		SELECT product_id, name, category, price
		FROM products
		WHERE product_id = $1
	`

	var p product.Product

	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Category,
		// &p.Description,
		&p.Price,
	// &p.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return product.Product{}, ErrNotFound
	}

	if err != nil {
		return product.Product{}, fmt.Errorf("find product by id: %w", err)
	}

	return p, nil
}

func (r *PostgreProductRepository) Save(ctx context.Context, p product.Product) {
}

func (r *PostgreProductRepository) Update(ctx context.Context, p product.Product) {
}

func (r *PostgreProductRepository) Delete(ctx context.Context, id uuid.UUID) {
}
