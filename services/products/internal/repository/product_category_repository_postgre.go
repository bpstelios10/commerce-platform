package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgreProductCategoryRepository struct {
	db       ProductCategoryDB
	timeouts Timeouts
}

type ProductCategoryDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewPostgreProductCategoryRepository(db ProductCategoryDB, t Timeouts) *PostgreProductCategoryRepository {
	return &PostgreProductCategoryRepository{db: db, timeouts: t}
}

func (r *PostgreProductCategoryRepository) Exists(ctx context.Context, category string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeouts.Read)
	defer cancel()

	const query = `
		SELECT name
		FROM product_categories
		WHERE name = $1
	`

	var name string

	err := r.db.QueryRow(ctx, query, category).Scan(&name)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("product category exists: %w", err)
	}

	return true, nil
}

func (r *PostgreProductCategoryRepository) List(ctx context.Context) ([]string, error) {
	// categories are very short so we can go with the lower timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeouts.Read)
	defer cancel()

	const query = `
		SELECT name
		FROM product_categories
	`

	categoriesNames := make([]string, 0)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return categoriesNames, fmt.Errorf("query product categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var categoryName string

		if err = rows.Scan(&categoryName); err != nil {
			return categoriesNames, fmt.Errorf("scan product categories: %w", err)
		}

		categoriesNames = append(categoriesNames, categoryName)
	}

	return categoriesNames, nil
}
