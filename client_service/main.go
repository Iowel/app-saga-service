package main

import (
	"clients/gapi"
	"clients/protos"
	repository "clients/reposiroty"
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const consumerGroup = "order_service"

func main() {
	db, err := repository.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	orderRepo := repository.NewOrderRepository(db)

	go runGrpcServer(orderRepo)

	go runGatewayServer(orderRepo)

	select {}
}

func runGrpcServer(orderRepo *repository.OrderRepository) {
	server, err := gapi.NewServer(orderRepo)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()

	protos.RegisterOrderServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":8086")
	if err != nil {
		log.Fatal("cannot launch gRPC server:", err)
	}

	log.Printf("Starting GRPC server on port %s\n", ":8086")
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot launch gRPC server:", err)
	}

}

func runGatewayServer(orderRepo *repository.OrderRepository) {
	const op = "order-service.RunGatewayServer"

	server, err := gapi.NewServer(orderRepo)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	// camel case for grpc
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = protos.RegisterOrderServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot create GatewayServer:", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/", grpcMux)

	alowedOrigins := []string{
		"http://158.160.30.125:8081",
		"http://158.160.30.125:8082",
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   alowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions, // важно!
		},
	})

	handler := c.Handler(mux)

	httpServer := &http.Server{
		Handler: handler,
		Addr:    "0.0.0.0:8087",
	}

	log.Printf("start HTTP gateway server at %s\n", httpServer.Addr)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Printf("HTTP gateway server failed to serve, path: %s, error: %v\n", op, err)
	}
}
