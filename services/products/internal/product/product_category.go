package product

import "strings"

type ProductCategory string

func (c ProductCategory) Normalize() ProductCategory {
	return ProductCategory(strings.ToUpper(strings.TrimSpace(string(c))))
}
