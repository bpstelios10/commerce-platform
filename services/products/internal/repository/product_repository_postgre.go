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
	db ProductDB
}

type ProductDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func NewPostgreProductRepository(db ProductDB) *PostgreProductRepository {
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

func (r *PostgreProductRepository) Save(ctx context.Context, p product.Product) error {
	// (product_id, name, category, description, price)
	const query = `
		INSERT INTO products (product_id, name, category, price)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, p.ID, p.Name, p.Category, p.Price)
	if err != nil {
		return fmt.Errorf("save product: %w", err)
	}

	return nil
}

func (r *PostgreProductRepository) Update(ctx context.Context, p product.Product) error {
	// (product_id, name, category, description, price)
	const query = `
		UPDATE products 
		SET name = $1, category = $2, price = $3
		WHERE product_id = $4
	`

	_, err := r.db.Exec(ctx, query, p.Name, p.Category, p.Price, p.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}

	return nil
}

func (r *PostgreProductRepository) Delete(ctx context.Context, id uuid.UUID) {
}
