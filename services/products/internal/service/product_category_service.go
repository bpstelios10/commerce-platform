package service

import (
	"context"
	"strings"
)

type ProductCategoryRepository interface {
	Exists(ctx context.Context, category string) bool
	GetAll(ctx context.Context) []string
}

type ProductCategoryService struct {
	repo ProductCategoryRepository
}

func NewProductCategoryService(repo ProductCategoryRepository) *ProductCategoryService {
	return &ProductCategoryService{repo: repo}
}

func (s *ProductCategoryService) Validate(ctx context.Context, category string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(category))
	if !s.repo.Exists(ctx, normalized) {
		return "", ErrInvalidCategory
	}

	return normalized, nil
}

func (s *ProductCategoryService) GetProductCategories(ctx context.Context) []string {
	return s.repo.GetAll(ctx)
}
