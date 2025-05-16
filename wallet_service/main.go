package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
	"wallet/kafka"
	"wallet/protos"
	"wallet/repository"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type WalletHandler struct {
	repo     *repository.BalanceRepository
	producer sarama.SyncProducer
}

const consumerGroup = "order_service"

func (w *WalletHandler) CheckBalance(ctx context.Context, message *sarama.ConsumerMessage) error {
	var product protos.OrderWithProduct

	if err := proto.Unmarshal(message.Value, &product); err != nil {
		return err
	}
	log.Printf("Check balance for order %d, user %d, price %d", product.Order.OrderID, product.Order.UserID, product.Product.Price)

	// Проверка и списание баланса
	err := w.repo.DeleteUserPrice(int(product.Product.Price), int(product.Order.UserID))

	balanceSufficient := err == nil
	if balanceSufficient {
		log.Printf("Balance sufficient for order %d", product.Order.OrderID)
	} else {
		log.Printf("Insufficient balance for order %d", product.Order.OrderID)
	}

	// Добавляем флаг в сообщение
	product.BalanceSufficient = balanceSufficient

	data, err := proto.Marshal(&product)
	if err != nil {
		return err
	}

	// Отправляем результат проверки баланса (решение о cancel/commit принимает оркестратор!)
	log.Printf("Sending balance_checked for order %d (balanceSufficient: %v)", product.Order.OrderID, balanceSufficient)
	if err := kafka.SendMessage(w.producer, "balance_checked", data); err != nil {
		log.Printf("Failed to send balance_checked message: %v", err)
		return err
	}

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

	handler := WalletHandler{
		repo: repository.NewBalanceRepository(db),
	}

	handler.producer, err = kafka.NewSyncProducer(brokers)
	if err != nil {
		log.Fatal(err)
	}

	if err := kafka.StartConsuming(ctx, brokers, "check_balance", consumerGroup, handler.CheckBalance); err != nil {
		log.Fatal(err)
	}

	select {}
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
