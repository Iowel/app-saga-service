package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Db struct {
	Pool *pgxpool.Pool
}

func NewDB() (*Db, error) {

	err := waitForPostgres("postgres://postgres:1234@postgres_product:5432/kafka_product?sslmode=disable", 10)
	if err != nil {
		log.Fatalf("Database unavailable: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), "postgres://postgres:1234@postgres_product:5432/kafka_product?sslmode=disable")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Connected to PostgreSQL")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS products (
		sku BIGINT PRIMARY KEY,
		price BIGINT NOT NULL,
		cnt BIGINT NOT NULL,
		avatar VARCHAR(255),
		name TEXT
	);
`

	_, err = pool.Exec(context.Background(), createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create products table: %v\n", err)
	}

	// Вставим дефолтные значения
	insertQuery := `
		--Продукты
	INSERT INTO products (sku, price, cnt, avatar, name) VALUES
		(1, 25, 67, 'static/ball.png', 'Мяч'),
		(2, 50, 78, 'static/snikers.png', 'Сникерс'),
		(3, 150, 134, 'static/phone.png', 'Телефон'),

		--Статусы
		(4, 1000, 60, 'static/status_gold.jpg', 'Золотой'),
		(5, 5000, 60, 'static/status_diamond.png', 'Бриллиантовый')
	ON CONFLICT (sku) DO NOTHING;
`

	_, err = pool.Exec(context.Background(), insertQuery)
	if err != nil {
		log.Fatalf("Failed to insert default products: %v\n", err)
	}

	return &Db{
		Pool: pool,
	}, nil

}

// для ожидания базы
func waitForPostgres(dsn string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		pool, err := pgxpool.New(context.Background(), dsn)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				log.Println("Successfully connected to postgres")
				pool.Close()
				return nil
			}
			pool.Close()
		}
		log.Printf("postgres not ready, retrying... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("could not connect to postgres after %d retries", maxRetries)
}
