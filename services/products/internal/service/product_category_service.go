package service

import "strings"

type ProductCategoryRepository interface {
	Exists(category string) bool
	GetAll() []string
}

type ProductCategoryService struct {
	repo ProductCategoryRepository
}

func NewProductCategoryService(repo ProductCategoryRepository) *ProductCategoryService {
	return &ProductCategoryService{repo: repo}
}

func (s *ProductCategoryService) Validate(category string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(category))
	if !s.repo.Exists(normalized) {
		return "", ErrInvalidCategory
	}

	return normalized, nil
}

func (s *ProductCategoryService) GetProductCategories() []string {
	return s.repo.GetAll()
}
