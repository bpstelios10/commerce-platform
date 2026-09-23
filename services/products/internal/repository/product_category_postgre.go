package repository

import (
	"context"
)

type PostgreProductCategoryRepository struct {
}

func NewPostgreProductCategoryRepository() *PostgreProductCategoryRepository {
	return &PostgreProductCategoryRepository{}
}

func (r *PostgreProductCategoryRepository) Exists(ctx context.Context, category string) bool {
	var found = false

	return found
}

func (r *PostgreProductCategoryRepository) GetAll(ctx context.Context) []string {
	var categoriesNames = make([]string, 0)

	return categoriesNames
}
