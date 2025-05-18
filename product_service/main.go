package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"product/gapi"
	"product/kafka"
	"product/protos"
	"product/repository"
	"time"

	"github.com/IBM/sarama"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type OrderHandler struct {
	repo     *repository.StockProductRepository
	producer sarama.SyncProducer
}

const consumerGroup = "order_service"

func (h *OrderHandler) CheckProduct(ctx context.Context, message *sarama.ConsumerMessage) error {
	var order protos.Order

	// get request
	if err := proto.Unmarshal(message.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %v", err)
	}
	log.Printf("Check OrderID: %d\n", order.OrderID)

	// Проверка наличия продукта
	product, err := h.repo.GetProduct(int(order.ProductSKU))
	if err != nil {
		log.Printf("Ошибка получения продукта: %v", err)
		return fmt.Errorf("db error: %w", err)
	}

	available := product.Cnt > 0
	if available {
		// Уменьшаем количество товара, если доступно
		if err := h.repo.DeleteProductCount(order.ProductSKU); err != nil {
			log.Printf("Failed to delete product count for SKU %d: %v", order.ProductSKU, err)
			return fmt.Errorf("failed to delete product count: %v", err)
		}
		log.Printf("Product %d reserved for order %d", order.ProductSKU, order.OrderID)
	} else {
		log.Printf("Товар %d закончился для order %d", order.ProductSKU, order.OrderID)
	}

	// Создаем сообщение с продуктом + флаг доступности
	orderWithProduct := &protos.OrderWithProduct{
		Order: &order,
		Product: &protos.Product{
			Sku:   product.Sku,
			Price: product.Price,
			Cnt:   product.Cnt,
			Name: product.Name,
		},
		Available: available,
	}

	// маршалим данные
	data, err := proto.Marshal(orderWithProduct)
	if err != nil {
		return err
	}

	// отправляем результат проверки в сервис оркестрации
	log.Printf("Sending product_checked for order %d (available: %v)", order.OrderID, available)
	if err := kafka.SendMessage(h.producer, "product_checked", data); err != nil {
		log.Printf("Failed to send product_checked message: %v", err)
	}

	return nil
}

func (h *OrderHandler) CancelWallet(ctx context.Context, message *sarama.ConsumerMessage) error {
	var product protos.OrderWithProduct

	if err := proto.Unmarshal(message.Value, &product); err != nil {
		return err
	}

	log.Printf("Rollback product count for ProductSKU %d (OrderID %d)", product.Order.ProductSKU, product.Order.OrderID)

	err := h.repo.BackProductCount(product.Order.ProductSKU)
	if err != nil {
		return err
	}
	log.Printf("Rollback done for order %d", product.Order.OrderID)

	return nil
}

func main() {

	err := waitForKafka("kafka:29092", 10)
	if err != nil {
		log.Fatal(err)
	}

	brokers := []string{"kafka:29092"}
	ctx := context.Background()

	db, err := repository.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Pool.Close()

	prodRepo := repository.NewStockProductRepository(db)
	go runGrpcServer(prodRepo)
	go runGatewayServer(prodRepo)

	handler := OrderHandler{
		repo: repository.NewStockProductRepository(db),
	}

	handler.producer, err = kafka.NewSyncProducer(brokers)
	if err != nil {
		log.Fatal(err)
	}

	if err := kafka.StartConsuming(ctx, brokers, "check_product", consumerGroup, handler.CheckProduct); err != nil {
		log.Fatal(err)
	}

	if err := kafka.StartConsuming(ctx, brokers, "cancel_wallet", consumerGroup, handler.CancelWallet); err != nil {
		log.Fatal(err)
	}

	select {}
}

func runGrpcServer(prodRepo *repository.StockProductRepository) {
	server, err := gapi.NewServer(prodRepo)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	grpcServer := grpc.NewServer()

	protos.RegisterOrderServiceServer(grpcServer, server)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":8089")
	if err != nil {
		log.Fatal("cannot launch gRPC server:", err)
	}

	log.Printf("Starting GRPC server on port %s\n", ":8089")
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot launch gRPC server:", err)
	}

}

func runGatewayServer(prodRepo *repository.StockProductRepository) {
	const op = "delivery.server.RunGatewayServer"

	server, err := gapi.NewServer(prodRepo)
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
		"http://84.201.150.225:8081",
		"http://84.201.150.225:8082",
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
		Addr:    "0.0.0.0:8088",
	}

	log.Printf("start HTTP gateway server at %s\n", httpServer.Addr)
	err = httpServer.ListenAndServe()
	if err != nil {
		log.Printf("HTTP gateway server failed to serve, path: %s, error: %v\n", op, err)
	}

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
