package product

import "github.com/google/uuid"

type Product struct {
	ID       uuid.UUID       `json:"id"`
	Name     string          `json:"name"`
	Category ProductCategory `json:"category"`
	Price    float64         `json:"price"`
	Stock    int             `json:"stock"`
}

func (p Product) DisplayName() string {
	return p.Name
}

func (p *Product) Rename(newName string) {
	p.Name = newName
}

func (p *Product) ApplyDiscount(percentage float64) {
	p.Price = p.Price * (1 - percentage/100)
}

func (p Product) IsExpensive() bool {
	return p.Price > 200
}
