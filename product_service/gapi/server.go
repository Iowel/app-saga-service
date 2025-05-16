package gapi

import (
	"product/protos"
	"product/repository"
)

type Server struct {
	protos.UnimplementedOrderServiceServer
	prodRepo *repository.StockProductRepository
}

func NewServer(prodRepo *repository.StockProductRepository) (*Server, error) {
	return &Server{
		prodRepo: prodRepo,
	}, nil
}
