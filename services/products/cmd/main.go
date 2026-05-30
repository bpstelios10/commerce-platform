package main

import (
	"commerce-platform/services/products/internal/product"
	"fmt"
)

func main() {
	product := product.Product{
		ID:    "1",
		Name:  "MacBook Pro",
		Price: 2500,
	}

	fmt.Println(product)
}
