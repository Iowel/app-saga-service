package repository

import (
	"context"
	"fmt"
	"time"
)

const (
	OrderCreated = iota
	OrderCommited
	OrderCanceled
)

type BalanceRepository struct {
	db *Db
}

func NewBalanceRepository(db *Db) *BalanceRepository {
	return &BalanceRepository{db: db}
}



func (u *BalanceRepository) DeleteUserPrice(price int, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Обновляем значение wallet и вычитаем price, если balance > 0
	query := `
		UPDATE profiles
		SET wallet = wallet - $1
		WHERE user_id = $2 AND wallet >= $1
		RETURNING wallet;`

	var newPrice int
	err := u.db.Pool.QueryRow(ctx, query, price, userId).Scan(&newPrice)
	if err != nil {
		return err
	}

	return nil
}



func (u *BalanceRepository) BackUserPrice(price int, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// обнлвляем значение wallet добавляя price
	query := `
		UPDATE profiles
		SET wallet = wallet + $1
		WHERE user_id = $2
		RETURNING wallet;`

	var newPrice int
	err := u.db.Pool.QueryRow(ctx, query, price, userId).Scan(&newPrice)
	if err != nil {
		return fmt.Errorf("failed to add count wallet: %v", err)
	}

	return nil
}
