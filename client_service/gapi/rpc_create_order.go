package gapi

import (
	"clients/kafka"
	"clients/protos"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppID       = errors.New("invalid app_id")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

func (server *Server) CreateOrder(ctx context.Context, req *protos.Order) (*protos.Order, error) {
	const op = "gapi.CreateOrder"

	// создаем заказ в репозитории
	order, err := server.orderRepo.CreateOrder(ctx, req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found: path; %s, err: %v", op, err)
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: path; %s, err: %v", op, err)
	}

	// формируем респонс для кафки
	kafkaResponse := &protos.Order{
		UserID:     order.UserID,
		ProductSKU: order.ProductSKU,
		OrderID:    order.OrderID,
	}

	err = waitForKafka("kafka:29092", 10)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Kafka unavailable: path; %s, err: %v", op, err)
	}

	brokers := []string{"kafka:29092"}

	syncProducer, err := kafka.NewSyncProducer(brokers)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create Kafka producer: path; %s, err: %v", op, err)
	}
	defer syncProducer.Close()


	// маршалим сообщения в протобуф
	msg, err := proto.Marshal(kafkaResponse)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal message: path; %s, err: %v", op, err)
	}

	// отправляем сообщение в кафку
	err = kafka.SendMessage(syncProducer, "create_order", msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send Kafka message: path; %s, err: %v", op, err)
	}

	//  ответ для gRPC
	gRPCResponse := &protos.Order{
		Status: "order submitted",
	}

	return gRPCResponse, nil
}

func waitForKafka(addr string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err == nil {
			_ = conn.Close()
			log.Println("Successfully connected to Kafka")
			return nil
		}
		log.Printf("Kafka not ready, retrying... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("could not connect to Kafka at %s", addr)
}
