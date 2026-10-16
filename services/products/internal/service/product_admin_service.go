package service

import (
	"commerce-platform/services/products/internal/product"
	"commerce-platform/services/products/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ProductCategoryService interface {
	Validate(ctx context.Context, category string) (string, error)
}

type ProductWriter interface {
	Create(ctx context.Context, p product.Product) (product.Product, error)
	Update(ctx context.Context, p product.Product) (product.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdminServiceImpl struct {
	productService  *ProductServiceImpl
	categoryService ProductCategoryService
	repo            ProductWriter
}

func NewAdminServiceImpl(productService *ProductServiceImpl, categoryService ProductCategoryService, repo ProductWriter) *AdminServiceImpl {
	return &AdminServiceImpl{productService: productService, categoryService: categoryService, repo: repo}
}

func (s *AdminServiceImpl) CreateProduct(ctx context.Context, name string, category string, description *string, price float64) (product.Product, error) {
	validatedCategory, err := s.categoryService.Validate(ctx, category)
	if err != nil {
		return product.Product{}, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return product.Product{}, fmt.Errorf("create product: %w", err)
	}

	p := product.Product{
		ID:          id,
		Name:        name,
		Category:    validatedCategory,
		Description: description,
		Price:       price,
	}

	logger := log(ctx)
	logger.Info().Str("product_id", p.ID.String()).Str("category", p.Category).Msg("creating product")

	p, err = s.repo.Create(ctx, p)
	if err != nil {
		return product.Product{}, fmt.Errorf("create product: %w", err)
	}

	return p, nil
}

func (s *AdminServiceImpl) UpdateProduct(
	ctx context.Context,
	id uuid.UUID,
	name string,
	category string,
	description *string,
	price float64,
	version int,
) (product.Product, error) {
	p, err := s.productService.GetProductByID(ctx, id)
	if err != nil {
		return product.Product{}, fmt.Errorf("updating product: %w", err)
	}

	validatedCategory, err := s.categoryService.Validate(ctx, category)
	if err != nil {
		return product.Product{}, fmt.Errorf("updating product: %w", err)
	}

	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("updating product")

	p.Name = name
	p.Category = validatedCategory
	p.Description = description
	p.Price = price
	p.Version = version

	p, err = s.repo.Update(ctx, p)
	if err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return product.Product{}, fmt.Errorf("updating product: %w", ErrProductUpdateConflict)
		}

		return product.Product{}, fmt.Errorf("updating product: %w", err)
	}

	return p, nil
}

func (s *AdminServiceImpl) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("attempting to delete product")

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}

	return nil
}
