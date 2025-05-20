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
