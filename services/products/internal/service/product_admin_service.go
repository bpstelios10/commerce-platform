package service

import (
	"commerce-platform/services/products/internal/product"
	"log/slog"
)

type ProductWriter interface {
	Save(product.Product)
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

	slog.Info(
		"creating product",
		"product", p,
	)

	s.repo.Save(p)
}
