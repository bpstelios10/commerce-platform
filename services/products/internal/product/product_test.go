package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyDiscount_WhenPercentageIsTen_UpdatesPrice(t *testing.T) {
	p := Product{Price: 200}

	p.ApplyDiscount(10)

	assert.InDelta(t, 180.0, p.Price, 0.000001)
}

func TestApplyDiscount_WhenPercentageIsZero_LeavesPriceUnchanged(t *testing.T) {
	p := Product{Price: 200}

	p.ApplyDiscount(0)

	assert.InDelta(t, 200.0, p.Price, 0.000001)
}
