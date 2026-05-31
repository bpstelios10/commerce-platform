package service

import (
	"commerce-platform/services/products/internal/product"
	"log/slog"
)

type ProductWriter interface {
	Save(product.Product)
	Update(product.Product)
	Delete(id string)
}

type AdminService struct {
	repo ProductWriter
}

func NewAdminService(repo ProductWriter) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) CreateProduct(id, name string, price float64) {
	p := product.Product{
		ID:    id,
		Name:  name,
		Price: price,
	}

	slog.Info("creating product", "product", p)

	s.repo.Save(p)
}

func (s *AdminService) UpdateProduct(id, name string, price float64) {
	slog.Info("updating product with", "productId", id)

	p := product.Product{
		ID:    id,
		Name:  name,
		Price: price,
	}

	s.repo.Update(p)
}

func (s *AdminService) DeleteProduct(id string) {
	slog.Info("attempting to delete product with", "productId", id)

	s.repo.Delete(id)
}
