package order

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid_WhenValidStatus_ReturnsTrue(t *testing.T) {
	validStatuses := []OrderStatus{CREATED, PAID, RETURNED, CANCELLED}

	for _, s := range validStatuses {
		assert.True(t, s.IsValid(), "expected %q to be valid", s)
	}
}

func TestIsValid_WhenInvalidStatus_ReturnsFalse(t *testing.T) {
	invalidStatuses := []OrderStatus{"", "paid", "PIAD", "UNKNOWN"}

	for _, s := range invalidStatuses {
		assert.False(t, s.IsValid(), "expected %q to be invalid", s)
	}
}
