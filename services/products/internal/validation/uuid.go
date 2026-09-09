package validation

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidUUID = errors.New("invalid UUID")
)

func GetValidUUID(id string) (uuid.UUID, error) {
	validUUID, err := uuid.Parse(id)
	if err != nil {
		return validUUID, ErrInvalidUUID
	}

	return validUUID, nil
}
