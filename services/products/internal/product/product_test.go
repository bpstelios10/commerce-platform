package product

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayName_WhenProductHasName_ReturnsName(t *testing.T) {
	p := Product{Name: "MacBook Pro"}

	assert.Equal(t, "MacBook Pro", p.DisplayName())
}

func TestRename_WhenNewNameProvided_UpdatesName(t *testing.T) {
	p := Product{Name: "Old Name"}

	p.Rename("New Name")

	assert.Equal(t, "New Name", p.Name)
}

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

func TestIsExpensive_WhenPriceAboveThreshold_ReturnsTrue(t *testing.T) {
	p := Product{Price: 200.01}

	assert.True(t, p.IsExpensive())
}

func TestIsExpensive_WhenPriceAtThresholdOrBelow_ReturnsFalse(t *testing.T) {
	assert.False(t, Product{Price: 200}.IsExpensive())
	assert.False(t, Product{Price: 199.99}.IsExpensive())
}
