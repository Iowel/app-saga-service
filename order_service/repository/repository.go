package repository

import (
	"context"
	"errors"
	"fmt"
	"order_service/model"
	"time"

	"github.com/jackc/pgx/v5"
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

// Update redis cache

func (u *OrderRepository) GetProfileByID(id int) (*model.Profile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select id, avatar, status, wallet, about from profiles where user_id = $1
	`

	row := u.db.Pool.QueryRow(ctx, query, id)

	var profile model.Profile

	err := row.Scan(&profile.ID, &profile.Avatar, &profile.Status, &profile.Wallet, &profile.About)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("profile with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	return &profile, nil
}

func (u *OrderRepository) GetUserByID(id int) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select id, email, password, name, avatar, role from users where id = $1
	`

	row := u.db.Pool.QueryRow(ctx, query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Avatar, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
