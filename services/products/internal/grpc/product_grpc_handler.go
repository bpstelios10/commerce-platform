package grpc

import (
	"commerce-platform/services/products/internal/service"
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
	p, err := h.service.GetProductByID(req.Id)
	if err != nil {
		return nil, HandleError(err)
	}

	return &GetProductByIDResponse{
		Id:    p.ID,
		Name:  p.Name,
		Price: p.Price,
	}, nil
}
