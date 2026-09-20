package grpc

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	context "context"
	"fmt"
)

type ProductGrpcHandler struct {
	UnimplementedProductServiceServer

	service *service.ProductService
}

func NewProductGrpcHandler(service *service.ProductService) *ProductGrpcHandler {
	return &ProductGrpcHandler{
		service: service,
	}
}

func (h *ProductGrpcHandler) GetProductByID(ctx context.Context, req *GetProductByIDRequest) (*GetProductByIDResponse, error) {
	logger := log(ctx)

	validUUID, err := validation.GetValidUUID(req.Id)
	if err != nil {
		return nil, HandleError(ctx, err)
	}

	p, err := h.service.GetProductByID(ctx, validUUID)
	if err != nil {
		return nil, HandleError(ctx, fmt.Errorf("get_product: %w", err))
	}

	logger.Info().Str("product_id", validUUID.String()).Msg("product was found")

	return &GetProductByIDResponse{
		Id:       p.ID.String(),
		Name:     p.Name,
		Category: p.Category,
		Price:    p.Price,
	}, nil
}
