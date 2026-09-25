package product

import "github.com/google/uuid"

type Product struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    float64   `json:"price"`
}

func (p *Product) ApplyDiscount(percentage float64) {
	p.Price = p.Price * (1 - percentage/100)
}
