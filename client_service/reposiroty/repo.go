package repository

import (
	"clients/protos"
	"context"
	"fmt"
)

type OrderRepository struct {
	db *Db
}

func NewOrderRepository(db *Db) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *protos.Order) (*protos.Order, error) {
	// начинаем транзакцию
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// вставляем заказ в таблицу orders
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, product_sku, timestamp, status, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING order_id
	`,
		order.UserID,
		order.ProductSKU,
		order.Timestamp,
		order.Status,
		order.Reason,
	).Scan(&order.OrderID)

	if err != nil {
		return nil, fmt.Errorf("failed insert order: %w", err)
	}

	// фиксируем транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed commit transaction: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) GetOrders(ctx context.Context, orderID int64) (*protos.Order, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT 
			order_id, 
			user_id, 
			timestamp, 
			product_sku, 
			status, 
			reason 
		FROM
			orders
		WHERE
			order_id = $1
	`, orderID)

	var order protos.Order
	err := row.Scan(
		&order.OrderID,
		&order.UserID,
		&order.Timestamp,
		&order.ProductSKU,
		&order.Status,
		&order.Reason,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context) ([]*protos.Order, error) {
	const query = `
        SELECT 
            user_id,
            timestamp,
            product_sku,
            order_id,
            status,
            reason
        FROM orders
    `

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []*protos.Order
	for rows.Next() {
		var order protos.Order
		if err := rows.Scan(
			&order.UserID,
			&order.Timestamp,
			&order.ProductSKU,
			&order.OrderID,
			&order.Status,
			&order.Reason,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*protos.Order, error) {
	const query = `
        SELECT 
            user_id,
            timestamp,
            product_sku,
            order_id,
            status,
            reason
        FROM orders
        WHERE user_id = $1
    `

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders by user_id: %w", err)
	}
	defer rows.Close()

	var orders []*protos.Order
	for rows.Next() {
		var order protos.Order
		if err := rows.Scan(
			&order.UserID,
			&order.Timestamp,
			&order.ProductSKU,
			&order.OrderID,
			&order.Status,
			&order.Reason,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) GetOrderRest(ctx context.Context, orderID int) (*Order, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT 
			order_id, 
			user_id, 
			timestamp, 
			product_sku, 
			status, 
			reason 
		FROM
			orders
		WHERE
			order_id = $1
	`, orderID)

	var order Order
	err := row.Scan(
		&order.OrderID,
		&order.UserID,
		&order.Timestamp,
		&order.ProductSKU,
		&order.Status,
		&order.Reason,
	)
	if err != nil {
		return nil, err
	}

	return &order, nil
}
