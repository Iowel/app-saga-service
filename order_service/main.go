package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"order_service/cache"
	"order_service/kafka"
	"order_service/model"
	"order_service/protos"
	"order_service/repository"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"google.golang.org/protobuf/proto"
)

type Orchestrator struct {
	producer  sarama.SyncProducer
	repo      *repository.OrderRepository
	cacheRepo cache.IPostCache
}

const consumerGroup = "order_service"

// отправляем заказ на проверку в базу продуктов
func (o *Orchestrator) StartSaga(ctx context.Context, message *sarama.ConsumerMessage) error {
	var order protos.Order

	if err := proto.Unmarshal(message.Value, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order: %v", err)
	}
	log.Printf("Starting saga for OrderID: %d\n", order.OrderID)

	if err := kafka.SendMessage(o.producer, "check_product", message.Value); err != nil {
		log.Printf("Failed to send check_product message: %v", err)
		return err
	}

	return nil
}

// обработка случая товара в наличии (если есть - отправляем на проверку баланса юзера, если товар закончился - отменяем заказ)
func (o *Orchestrator) ProcessProductChecked(ctx context.Context, message *sarama.ConsumerMessage) error {
	var product protos.OrderWithProduct
	if err := proto.Unmarshal(message.Value, &product); err != nil {
		return err
	}

	if !product.Available {
		log.Printf("Product unavailable for OrderID: %d, cancelling order", product.Order.OrderID)

		_ = o.repo.UpdateReason(ctx, product.Order.OrderID, "Товар закончился")

		_ = kafka.SendMessage(o.producer, "cancel_order", message.Value)

		return nil
	}

	log.Printf("Product available for OrderID: %d, checking balance", product.Order.OrderID)
	_ = kafka.SendMessage(o.producer, "check_balance", message.Value)
	return nil
}

// обработка случая наличия денег у юзера, если денег нет - отменяем транзакцию, если есть - коммитим
func (o *Orchestrator) ProcessBalanceChecked(ctx context.Context, message *sarama.ConsumerMessage) error {
	var product protos.OrderWithProduct
	if err := proto.Unmarshal(message.Value, &product); err != nil {
		return err
	}

	if !product.BalanceSufficient {
		_ = o.repo.UpdateReason(ctx, product.Order.OrderID, "Недостаточно средств")
		log.Printf("Balance insufficient for OrderID: %d, cancelling order", product.Order.OrderID)

		// обновляем статус заказа
		o.repo.UpdateStatus(ctx, product.Order.OrderID, "cancel")

		_ = kafka.SendMessage(o.producer, "cancel_wallet", message.Value)

		return nil
	}

	log.Printf("Balance sufficient for OrderID: %d, committing order", product.Order.OrderID)
	_ = kafka.SendMessage(o.producer, "commit_order", message.Value)
	return nil
}

func (o *Orchestrator) ProcessCommitOrder(ctx context.Context, message *sarama.ConsumerMessage) error {
	var product protos.OrderWithProduct

	if err := proto.Unmarshal(message.Value, &product); err != nil {
		return err
	}
	log.Printf("OrderID %d successfully committed!", product.Order.OrderID)

	// обновляем статус профиля
	// первые 3 сущности ето продукты, а после 3 идут статусы, их и обновляем
	if product.Product.Sku > 3 {
		productName := product.Product.Name
		log.Printf("productName %s", productName)

		userID := product.Order.UserID
		log.Printf("userID %d", userID)

		// меняем имя статуса у юзера на приобретенный
		if err := o.repo.UpdateUserStatus(ctx, productName, userID); err != nil {
			log.Printf("Failed to update user status: %v", err)
		}

		// Update redis cache
		newProfile, _ := o.repo.GetProfileByID(int(userID))
		newUser, _ := o.repo.GetUserByID(int(userID))

		var cacheUser = &model.UserCache{
			ID:        newUser.ID,
			Email:     newUser.Email,
			Password:  newUser.Password,
			Name:      newUser.Name,
			Role:      newUser.Role,
			Avatar:    newProfile.Avatar,
			Status:    newProfile.Status,
			Wallet:    newProfile.Wallet,
			CreatedAt: newUser.CreatedAt,
			UpdatedAt: newUser.UpdatedAt,
		}

		idStr := "user:" + strconv.Itoa(newUser.ID)
		o.cacheRepo.Set(idStr, cacheUser)
	}

	// обновляем статус заказа
	if err := o.repo.UpdateStatus(ctx, product.Order.OrderID, "success"); err != nil {
		log.Printf("Failed to update order status: %v", err)
	}

	// обновляем причину
	if err := o.repo.UpdateReason(ctx, product.Order.OrderID, "Товар успешно оплачен"); err != nil {
		log.Printf("Failed to update order reason: %v", err)
	}

	// упаковываем и отправляем в хендлер для клиента
	orderWithProduct := &protos.OrderWithProduct{
		Order:   product.Order,
		Product: product.Product,
	}
	data, err := proto.Marshal(orderWithProduct)
	if err != nil {
		return err
	}
	if err := kafka.SendMessage(o.producer, "get_product", data); err != nil {
		log.Printf("Failed to send product_checked message: %v", err)
	}

	return nil
}

// помечаеи заказ как отмененный
func (o *Orchestrator) CancelOrder(ctx context.Context, message *sarama.ConsumerMessage) error {
	var product protos.OrderWithProduct
	if err := proto.Unmarshal(message.Value, &product); err != nil {
		return err
	}

	// обновляем статус заказа
	o.repo.UpdateStatus(ctx, product.Order.OrderID, "cancel")

	return nil
}

func main() {

	err := waitForKafka("kafka:29092", 10)
	if err != nil {
		log.Fatal(err)
	}

	// kafka
	brokers := []string{"kafka:29092"}
	ctx := context.Background()

	producer, err := kafka.NewSyncProducer(brokers)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// db
	db, err := repository.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Pool.Close()

	// redis
	redisCache := cache.NewRedisCache("redis:6379", 1, 99999999999)

	// init Orchestrator
	orc := Orchestrator{
		producer:  producer,
		repo:      repository.NewOrderRepository(db),
		cacheRepo: redisCache,
	}

	go func() {
		if err := kafka.StartConsuming(ctx, brokers, "create_order", consumerGroup, orc.StartSaga); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := kafka.StartConsuming(ctx, brokers, "product_checked", consumerGroup, orc.ProcessProductChecked); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := kafka.StartConsuming(ctx, brokers, "balance_checked", consumerGroup, orc.ProcessBalanceChecked); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := kafka.StartConsuming(ctx, brokers, "commit_order", consumerGroup, orc.ProcessCommitOrder); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := kafka.StartConsuming(ctx, brokers, "cancel_order", consumerGroup, orc.CancelOrder); err != nil {
			log.Fatal(err)
		}
	}()

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
