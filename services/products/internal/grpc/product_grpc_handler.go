package grpc

import (
	"commerce-platform/services/products/internal/service"
	"commerce-platform/services/products/internal/validation"
	context "context"
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
	validUUID, err := validation.GetValidUUID(req.Id)
	if err != nil {
		return nil, HandleError(err)
	}

	p, err := h.service.GetProductByID(validUUID)
	if err != nil {
		return nil, HandleError(err)
	}

	return &GetProductByIDResponse{
		Id:       p.ID.String(),
		Name:     p.Name,
		Category: p.Category,
		Price:    p.Price,
	}, nil
}
