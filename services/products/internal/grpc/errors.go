package grpc

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO add logs here
func HandleError(err error) error {

	switch {

	case errors.Is(err, service.ErrProductNotFound):
		return status.Error(
			codes.NotFound,
			service.ErrProductNotFound.Error(),
		)

	case errors.Is(err, service.ErrInvalidProduct):
		return status.Error(
			codes.InvalidArgument,
			service.ErrInvalidProduct.Error(),
		)

	case errors.Is(err, service.ErrInvalidCategory):
		return status.Error(
			codes.InvalidArgument,
			service.ErrInvalidCategory.Error(),
		)

	case errors.Is(err, validation.ErrInvalidUUID):
		return status.Error(
			codes.InvalidArgument,
			validation.ErrInvalidUUID.Error(),
		)

	default:
		return status.Error(
			codes.Internal,
			"internal server error",
		)
	}
}
