package gapi

import (
	"context"
	"fmt"

	"clients/protos"
)

func (server *Server) GetAllOrders(ctx context.Context, req *protos.GetAllOrdersRequest) (*protos.GetAllOrdersResponse, error) {
	const op = "gapi.GetOrder"

	orders, err := server.orderRepo.GetAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get products: %w", op, err)
	}

	return &protos.GetAllOrdersResponse{
		Orders: orders,
	}, nil

}

func (server *Server) GetOrdersByUser(ctx context.Context, req *protos.GetOrdersByUserRequest) (*protos.GetOrdersByUserResponse, error) {
	const op = "gapi.GetOrder"

	orders, err := server.orderRepo.GetOrdersByUserID(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get products: %w", op, err)
	}

	return &protos.GetOrdersByUserResponse{
		Orders: orders,
	}, nil

}
