package service

import (
	"context"
	"fmt"
	"strings"
)

type ProductCategoryRepository interface {
	Exists(ctx context.Context, category string) (bool, error)
	GetAll(ctx context.Context) ([]string, error)
}

type ProductCategoryServiceImpl struct {
	repo ProductCategoryRepository
}

func NewProductCategoryServiceImpl(repo ProductCategoryRepository) *ProductCategoryServiceImpl {
	return &ProductCategoryServiceImpl{repo: repo}
}

func (s *ProductCategoryServiceImpl) Validate(ctx context.Context, category string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(category))

	exists, err := s.repo.Exists(ctx, normalized)
	if err != nil {
		return "", fmt.Errorf("validate product category: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("validate product category %s: %w", normalized, ErrInvalidCategory)
	}

	return normalized, nil
}

// TODO add category as domain, including the description
func (s *ProductCategoryServiceImpl) GetProductCategories(ctx context.Context) ([]string, error) {
	categories, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get product categories: %w", err)
	}

	return categories, nil
}
