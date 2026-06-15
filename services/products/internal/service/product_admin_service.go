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
	productService *ProductService
	repo           ProductWriter
}

func NewAdminService(productService *ProductService, repo ProductWriter) *AdminService {
	return &AdminService{productService: productService, repo: repo}
}

func (s *AdminService) CreateProduct(name string, price float64) (product.Product, error) {
	id, _ := uuid.NewV7()
	p := product.Product{
		ID:    id,
		Name:  name,
		Price: price,
	}

	slog.Info("creating product", "product", p)

	s.repo.Save(p)

	return p, nil
}

func (s *AdminService) UpdateProduct(id uuid.UUID, name string, price float64) (product.Product, error) {
	if _, err := s.productService.GetProductByID(id); err != nil {
		return product.Product{}, err
	}

	slog.Info("updating product with", "productId", id)

	p := product.Product{
		ID:    id,
		Name:  name,
		Price: price,
	}

	s.repo.Update(p)
	return p, nil
}

func (s *AdminService) DeleteProduct(id uuid.UUID) {
	slog.Info("attempting to delete product with", "productId", id)

	s.repo.Delete(id)
}
