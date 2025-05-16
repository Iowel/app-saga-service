package gapi

import (
	"context"
	"log"

	"clients/protos"
)

func (server *Server) GetOrder(ctx context.Context, req *protos.OrderRequest) (*protos.Order, error) {
	const op = "gapi.GetOrder"

	log.Println(req.OrderId)

	order, err := server.orderRepo.GetOrders(ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return order, nil

}
