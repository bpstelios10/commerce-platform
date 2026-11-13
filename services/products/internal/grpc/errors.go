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
	var notFoundErr *service.ErrProductNotFound

	switch {

	case errors.As(err, &notFoundErr):
		logger.Warn().Err(err).Msg("product not found")
		return status.Error(codes.NotFound, notFoundErr.Error())

	// will never happen for get product by id
	// case errors.Is(err, service.ErrInvalidProduct):
	// 	logger.Warn().Err(err).Msg("invalid product")
	// 	return status.Error(codes.InvalidArgument, service.ErrInvalidProduct.Error())

	// case errors.Is(err, service.ErrInvalidCategory):
	// 	logger.Warn().Err(err).Msg("invalid product category")
	// 	return status.Error(codes.InvalidArgument, service.ErrInvalidCategory.Error())

	case errors.Is(err, validation.ErrInvalidUUID):
		logger.Warn().Err(err).Msg("invalid UUID")
		return status.Error(codes.InvalidArgument, validation.ErrInvalidUUID.Error())

	case errors.Is(err, context.DeadlineExceeded):
		logger.Warn().Err(err).Msg("request timed out")
		return status.Error(codes.DeadlineExceeded, "request timed out")

	case errors.Is(err, context.Canceled):
		// debug cause this is normal client behavior, not a server fault
		logger.Debug().Err(err).Msg("request canceled by client")
		return status.Error(codes.Canceled, "request canceled by client")

	default:
		logger.Error().Err(err).Msg("unexpected error handled, with")
		return status.Error(codes.Internal, "internal server error")
	}
}
