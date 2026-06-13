package validation

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

var (
	ErrInvalidUUID = errors.New("invalid UUID")
)

func GetValidUUID(id string) (uuid.UUID, error) {
	validUUID, err := uuid.Parse(id)
	if err != nil {
		slog.Warn("not a valid uuid format for", "id", id)
		return validUUID, ErrInvalidUUID
	}

	return validUUID, nil
}
