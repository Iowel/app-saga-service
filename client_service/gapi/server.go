package gapi

import (
	repository "clients/reposiroty"

	"clients/protos"
)

type Server struct {
	protos.UnimplementedOrderServiceServer
	orderRepo *repository.OrderRepository
}

func NewServer(orderRepo *repository.OrderRepository) (*Server, error) {
	return &Server{
		orderRepo: orderRepo,
	}, nil
}
