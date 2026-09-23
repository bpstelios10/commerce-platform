package repository

import (
	"context"

	"commerce-platform/services/products/internal/product"

	"github.com/google/uuid"
)

type PostgreProductRepository struct {
	db ProductCategoryDB
}

type ProductDB interface {
}

func NewPostgreProductRepository(db ProductCategoryDB) *PostgreProductRepository {
	return &PostgreProductRepository{db: db}
}

func (r *PostgreProductRepository) FindAll(ctx context.Context) []product.Product {
	var products []product.Product

	return products
}

func (r *PostgreProductRepository) FindByID(ctx context.Context, id uuid.UUID) (product.Product, bool) {
	return product.Product{}, false
}

func (r *PostgreProductRepository) Save(ctx context.Context, p product.Product) {
}

func (r *PostgreProductRepository) Update(ctx context.Context, p product.Product) {
}

func (r *PostgreProductRepository) Delete(ctx context.Context, id uuid.UUID) {
}
