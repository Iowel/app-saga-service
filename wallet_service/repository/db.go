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

	createProfilesTableQuery := `
	CREATE TABLE IF NOT EXISTS profiles (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL UNIQUE,
		avatar VARCHAR(255),
		about TEXT,
		friends INTEGER[],
		created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
		wallet INTEGER DEFAULT 0
	);
`

	_, err = pool.Exec(context.Background(), createProfilesTableQuery)
	if err != nil {
		log.Fatalf("Failed to create profiles table: %v\n", err)
	}

	// Вставим дефолтные значения
	insertQuery := `
		INSERT INTO profiles (user_id, avatar, friends, wallet) VALUES
			(1, 'static/ball.png', '{}', 500)
		ON CONFLICT (user_id) DO NOTHING;
		`

	_, err = pool.Exec(context.Background(), insertQuery)
	if err != nil {
		log.Fatalf("Failed to insert default profiles: %v\n", err)
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
