package service

import (
	"commerce-platform/services/products/internal/product"
	"context"

	"github.com/google/uuid"
)

type ProductWriter interface {
	Save(ctx context.Context, p product.Product)
	Update(ctx context.Context, p product.Product)
	Delete(ctx context.Context, id uuid.UUID)
}

type AdminService struct {
	productService  *ProductService
	categoryService *ProductCategoryService
	repo            ProductWriter
}

func NewAdminService(productService *ProductService, categoryService *ProductCategoryService, repo ProductWriter) *AdminService {
	return &AdminService{productService: productService, categoryService: categoryService, repo: repo}
}

func (s *AdminService) CreateProduct(ctx context.Context, name string, category string, price float64, stock int) (product.Product, error) {
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
		Stock:    stock,
	}

	logger := log(ctx)
	logger.Info().Str("product_id", p.ID.String()).Str("category", p.Category).Msg("creating product")

	s.repo.Save(ctx, p)

	return p, nil
}

func (s *AdminService) UpdateProduct(ctx context.Context, id uuid.UUID, name string, category string, price float64, stock int) (product.Product, error) {
	if _, err := s.productService.GetProductByID(ctx, id); err != nil {
		return product.Product{}, err
	}

	validatedCategory, err := s.categoryService.Validate(ctx, category)
	if err != nil {
		return product.Product{}, err
	}

	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("updating product")

	p := product.Product{
		ID:       id,
		Name:     name,
		Category: validatedCategory,
		Price:    price,
		Stock:    stock,
	}

	s.repo.Update(ctx, p)
	return p, nil
}

func (s *AdminService) DeleteProduct(ctx context.Context, id uuid.UUID) {
	logger := log(ctx)
	logger.Info().Str("product_id", id.String()).Msg("attempting to delete product")

	s.repo.Delete(ctx, id)
}
