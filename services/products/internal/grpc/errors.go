package grpc

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleError(ctx context.Context, err error) error {
	logger := log(ctx)

	switch {

	case errors.Is(err, service.ErrProductNotFound):
		logger.Warn().Err(err).Msg("product not found")
		return status.Error(
			codes.NotFound,
			service.ErrProductNotFound.Error(),
		)

	case errors.Is(err, service.ErrInvalidProduct):
		logger.Warn().Err(err).Msg("invalid product")
		return status.Error(
			codes.InvalidArgument,
			service.ErrInvalidProduct.Error(),
		)

	case errors.Is(err, service.ErrInvalidCategory):
		logger.Warn().Err(err).Msg("invalid product category")
		return status.Error(
			codes.InvalidArgument,
			service.ErrInvalidCategory.Error(),
		)

	case errors.Is(err, validation.ErrInvalidUUID):
		logger.Warn().Err(err).Msg("invalid UUID")
		return status.Error(
			codes.InvalidArgument,
			validation.ErrInvalidUUID.Error(),
		)

	default:
		logger.Error().Err(err).Msg("unexpected error handled, with")
		return status.Error(
			codes.Internal,
			"internal server error",
		)
	}
}
