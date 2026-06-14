package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidUUID_WhenValidUUID(t *testing.T) {
	id, err := GetValidUUID("f47ac10b-58cc-4372-a567-0e02b2c3d001")

	assert.NoError(t, err)
	assert.NotNil(t, id)
}

func TestGetValidUUID_WhenInvalidUUID_returnsError(t *testing.T) {
	id, err := GetValidUUID("f47ac10b-58cc-4372-a567-0e02b")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidUUID)
	assert.Empty(t, id)
}
