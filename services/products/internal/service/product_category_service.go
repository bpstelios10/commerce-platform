package service

import (
	"context"
	"fmt"
	"strings"
)

type ProductCategoryRepository interface {
	Exists(ctx context.Context, category string) (bool, error)
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

	exists, err := s.repo.Exists(ctx, normalized)
	if err != nil {
		return "", fmt.Errorf("validate product category: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("validate product category %s: %w", normalized, ErrInvalidCategory)
	}

	return normalized, nil
}

func (s *ProductCategoryService) GetProductCategories(ctx context.Context) []string {
	return s.repo.GetAll(ctx)
}
