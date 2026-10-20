package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyDiscount_WhenPercentageIsTen_UpdatesPrice(t *testing.T) {
	p := Product{Price: 200}

	err := p.ApplyDiscount(10)

	assert.NoError(t, err)
	assert.InDelta(t, 180.0, p.Price, 0.000001)
}

func TestApplyDiscount_WhenPercentageIsLessThan1_ReturnsError(t *testing.T) {
	p := Product{Price: 200}

	err := p.ApplyDiscount(0)

	assert.Error(t, err)
	assert.EqualError(t, err, "Discount amount should be between 1 and 99")
	assert.InDelta(t, 200.0, p.Price, 0.000001)
}

func TestApplyDiscount_WhenPercentageMoreThan99_ReturnsError(t *testing.T) {
	p := Product{Price: 200}

	err := p.ApplyDiscount(100)

	assert.Error(t, err)
	assert.EqualError(t, err, "Discount amount should be between 1 and 99")
	assert.InDelta(t, 200.0, p.Price, 0.000001)
}
