package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const dsn string = "postgres://postgres:1234@postgres_go:5432/authservice?sslmode=disable"

type Db struct {
	Pool *pgxpool.Pool
}

func NewDB() (*Db, error) {

	err := waitForPostgres(dsn, 10)
	if err != nil {
		log.Fatalf("Database unavailable: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	log.Println("Connected to PostgreSQL")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS orders (
		order_id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		product_sku BIGINT NOT NULL,
		timestamp BIGINT NOT NULL,
		status TEXT,
		reason TEXT
	);
  `

	_, err = pool.Exec(context.Background(), createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}

	return &Db{
		Pool: pool,
	}, nil

}

// ждем загрузки базы
func waitForPostgres(dsn string, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		pool, err := pgxpool.New(context.Background(), dsn)
		if err == nil {
			err = pool.Ping(context.Background())
			if err == nil {
				log.Println("Successfully connected to PostgreSQL")
				pool.Close()
				return nil
			}
			pool.Close()
		}
		log.Printf("PostgreSQL not ready, retrying... (%d/%d)\n", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("could not connect to PostgreSQL after %d retries", maxRetries)
}
