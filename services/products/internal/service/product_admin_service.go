package service

import (
	"commerce-platform/services/products/internal/product"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type ProductWriter interface {
	Save(ctx context.Context, p product.Product) error
	Update(ctx context.Context, p product.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AdminService struct {
	productService  *ProductService
	categoryService *ProductCategoryService
	repo            ProductWriter
}

func NewAdminService(productService *ProductService, categoryService *ProductCategoryService, repo ProductWriter) *AdminService {
	return &AdminService{productService: productService, categoryService: categoryService, repo: repo}
}

func (s *AdminService) CreateProduct(ctx context.Context, name string, category string, price float64) (product.Product, error) {
	validatedCategory, err := s.categoryService.Validate(ctx, category)
	if err != nil {
		return product.Product{}, err
	}

	id, _ := uuid.NewV7()
	p := product.Product{
		ID:       id,
		Name:     name,
		Category: validatedCategory,
		Price:    price,
	}

	logger := log(ctx)
	logger.Info().Str("product_id", p.ID.String()).Str("category", p.Category).Msg("creating product")

	err = s.repo.Save(ctx, p)
	if err != nil {
		return product.Product{}, fmt.Errorf("create product: %w", err)
	}

	return p, nil
}

func (s *AdminService) UpdateProduct(ctx context.Context, id uuid.UUID, name string, category string, price float64) (product.Product, error) {
	if _, err := s.productService.GetProductByID(ctx, id); err != nil {
		return product.Product{}, fmt.Errorf("updating product: %w", err)
	}

	validatedCategory, err := s.categoryService.Validate(ctx, category)
	if err != nil {
		return product.Product{}, fmt.Errorf("updating product: %w", err)
	}

	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("updating product")

	p := product.Product{
		ID:       id,
		Name:     name,
		Category: validatedCategory,
		Price:    price,
	}

	err = s.repo.Update(ctx, p)
	if err != nil {
		return product.Product{}, fmt.Errorf("updating product: %w", err)
	}

	return p, nil
}

func (s *AdminService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("attempting to delete product")

	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}

	return nil
}
