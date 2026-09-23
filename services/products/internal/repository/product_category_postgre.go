package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgreProductCategoryRepository struct {
	db DB
}

type DB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewPostgreProductCategoryRepository(db DB) *PostgreProductCategoryRepository {
	return &PostgreProductCategoryRepository{db: db}
}

func (r *PostgreProductCategoryRepository) Exists(ctx context.Context, category string) (bool, error) {
	const query = `
		SELECT name
		FROM product_categories
		WHERE name ILIKE $1
	`

	var name string

	err := r.db.QueryRow(ctx, query, category).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("product category exists: %w", err)
	}

	return true, nil
}

func (r *PostgreProductCategoryRepository) GetAll(ctx context.Context) ([]string, error) {
	const query = `
		SELECT name
		FROM product_categories
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query product categories: %w", err)
	}
	defer rows.Close()

	var categoriesNames []string

	for rows.Next() {
		var categoryName string

		if err = rows.Scan(&categoryName); err != nil {
			return nil, fmt.Errorf("scan product categories: %w", err)
		}

		categoriesNames = append(categoriesNames, categoryName)
	}

	return categoriesNames, nil
}
