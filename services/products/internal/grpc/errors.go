package grpc

import (
	"commerce-platform/services/products/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleError(err error) error {

	switch err {

	case service.ErrProductNotFound:
		return status.Error(
			codes.NotFound,
			err.Error(),
		)

	case service.ErrInvalidProduct:
		return status.Error(
			codes.InvalidArgument,
			err.Error(),
		)

	default:
		return status.Error(
			codes.Internal,
			"internal server error",
		)
	}
}
