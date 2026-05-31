package service

import (
	"commerce-platform/services/products/internal/product"
)

type ProductRepository interface {
	FindAll() []product.Product
	FindByID(id string) (product.Product, bool)
}

type ProductService struct {
	repository ProductRepository
}

func NewProductService(repository ProductRepository) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) GetProducts() []product.Product {
	return s.repository.FindAll()
}

func (s *ProductService) GetProductByID(id string) (product.Product, bool) {
	return s.repository.FindByID(id)
}
