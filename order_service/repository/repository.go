package repository

import (
	"context"
	"fmt"
)

const (
	OrderCreated = iota
	OrderCommited
	OrderCanceled
)

type OrderRepository struct {
	db *Db
}

func NewOrderRepository(db *Db) *OrderRepository {
	return &OrderRepository{db: db}
}

func (u *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
	query := `UPDATE orders SET status = $1 WHERE order_id = $2`

	_, err := u.db.Pool.Exec(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("failed to update status for order %d: %w", orderID, err)
	}

	return nil
}

func (u *OrderRepository) UpdateReason(ctx context.Context, orderID int64, reason string) error {
	query := `UPDATE orders SET reason = $1 WHERE order_id = $2`

	_, err := u.db.Pool.Exec(ctx, query, reason, orderID)
	if err != nil {
		return fmt.Errorf("failed to update reason for order %d: %w", orderID, err)
	}

	return nil
}

func (u *OrderRepository) UpdateUserStatus(ctx context.Context, status string, userID int64) error {
	query := `
		UPDATE profiles 
		SET status = $1 
		WHERE user_id = $2
	`
	_, err := u.db.Pool.Exec(ctx, query, status, userID)
	if err != nil {
		return fmt.Errorf("failed to update status for order %d: %w", userID, err)
	}
	return nil
}
