package product

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description *string   `json:"description,omitempty"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	Version     int       `json:"version"`
}

func (p *Product) ApplyDiscount(percentage int16) error {
	if percentage < 1 || percentage > 99 {
		return errors.New("Discount amount should be between 1 and 99")
	}

	p.Price = p.Price * (1 - float64(percentage)/100)

	return nil
}
