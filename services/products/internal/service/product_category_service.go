package service

import "commerce-platform/services/products/internal/product"

type ProductCategoryRepository interface {
	Exists(category product.ProductCategory) bool
}

type ProductCategoryService struct {
	repo ProductCategoryRepository
}

func NewProductCategoryService(repo ProductCategoryRepository) *ProductCategoryService {
	return &ProductCategoryService{repo: repo}
}

func (s *ProductCategoryService) Validate(category product.ProductCategory) error {
	if !s.repo.Exists(category) {
		return ErrInvalidCategory
	}
	return nil
}
