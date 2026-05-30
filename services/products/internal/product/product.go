package product

type Product struct {
	ID    string
	Name  string
	Price float64
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
