package gapi

import (
	"context"
	"fmt"
	"product/protos"
)

func (server *Server) GetAllProducts(ctx context.Context, req *protos.GetAllProductsRequest) (*protos.GetAllProductsResponse, error) {
	const op = "product_service.GetAllProducts"

	products, err := server.prodRepo.GetAllProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get products: %w", op, err)
	}

	return &protos.GetAllProductsResponse{
		Products: products,
	}, nil
}
