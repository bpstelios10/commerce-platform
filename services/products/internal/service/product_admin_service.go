package service

import (
	"commerce-platform/services/products/internal/product"
	"log/slog"

	"github.com/google/uuid"
)

type ProductWriter interface {
	Save(product.Product)
	Update(product.Product)
	Delete(id uuid.UUID)
}

type AdminService struct {
	productService  *ProductService
	categoryService *ProductCategoryService
	repo            ProductWriter
}

func NewAdminService(productService *ProductService, categoryService *ProductCategoryService, repo ProductWriter) *AdminService {
	return &AdminService{productService: productService, categoryService: categoryService, repo: repo}
}

func (s *AdminService) CreateProduct(name string, category product.ProductCategory, price float64, stock int) (product.Product, error) {
	category = category.Normalize()
	if err := s.categoryService.Validate(category); err != nil {
		return product.Product{}, err
	}

	id, _ := uuid.NewV7()
	p := product.Product{
		ID:       id,
		Name:     name,
		Category: category,
		Price:    price,
		Stock:    stock,
	}

	slog.Info("creating product", "product", p)

	s.repo.Save(p)

	return p, nil
}

func (s *AdminService) UpdateProduct(id uuid.UUID, name string, category product.ProductCategory, price float64, stock int) (product.Product, error) {
	if _, err := s.productService.GetProductByID(id); err != nil {
		return product.Product{}, err
	}

	category = category.Normalize()
	if err := s.categoryService.Validate(category); err != nil {
		return product.Product{}, err
	}

	slog.Info("updating product with", "productId", id)

	p := product.Product{
		ID:       id,
		Name:     name,
		Category: category,
		Price:    price,
		Stock:    stock,
	}

	s.repo.Update(p)
	return p, nil
}

func (s *AdminService) DeleteProduct(id uuid.UUID) {
	slog.Info("attempting to delete product with", "productId", id)

	s.repo.Delete(id)
}
