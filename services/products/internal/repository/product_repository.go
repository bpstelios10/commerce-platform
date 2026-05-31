package repository

import "commerce-platform/services/products/internal/product"

type InMemoryProductRepository struct {
	products map[string]product.Product
}

func NewInMemoryProductRepository() *InMemoryProductRepository {
	return &InMemoryProductRepository{
		products: map[string]product.Product{
			"1": {
				ID:    "1",
				Name:  "MacBook Pro",
				Price: 2500,
			},
			"2": {
				ID:    "2",
				Name:  "iPhone",
				Price: 1200,
			},
		},
	}
}

func (r *InMemoryProductRepository) FindAll() []product.Product {
	var products []product.Product

	for _, p := range r.products {
		products = append(products, p)
	}

	return products
}

func (r *InMemoryProductRepository) FindByID(id string) (product.Product, bool) {
	p, found := r.products[id]
	return p, found
}
