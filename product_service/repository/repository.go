package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"product/protos"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	OrderCreated = iota
	OrderCommited
	OrderCanceled
)

type StockProductRepository struct {
	db *Db
}

func NewStockProductRepository(db *Db) *StockProductRepository {
	return &StockProductRepository{db: db}
}

func (s *StockProductRepository) GetProduct(id int) (*protos.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query :=
		`SELECT
		sku, price, cnt, name
	FROM
		products
	WHERE
		sku = $1`

	row := s.db.Pool.QueryRow(ctx, query, id)

	var product protos.Product
	err := row.Scan(&product.Sku, &product.Price, &product.Cnt, &product.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("product with id %d not found", id)
			return nil, err

		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (u *StockProductRepository) DeleteProductCount(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE
			products
		SET
			cnt = cnt - 1
		WHERE
			sku = $1 AND cnt > 0
		RETURNING
			cnt;`

	var newCount int64
	err := u.db.Pool.QueryRow(ctx, query, id).Scan(&newCount)
	if err != nil {
		return fmt.Errorf("failed to delete count: %v", err)
	}

	return nil
}

func (u *StockProductRepository) BackProductCount(sku int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE
			products
		SET
			cnt = cnt + 1
		WHERE
			sku = $1 AND cnt > 0
		RETURNING
			cnt;`

	var newCount int64
	err := u.db.Pool.QueryRow(ctx, query, sku).Scan(&newCount)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no product found with sku %d", sku)
		}
		return fmt.Errorf("failed to back count: %v", err)
	}

	return nil
}

func (u *StockProductRepository) GetAllProducts(ctx context.Context) ([]*protos.Product, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `SELECT sku, price, cnt, avatar, name FROM products`

	rows, err := u.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*protos.Product
	for rows.Next() {
		var p protos.Product
		err := rows.Scan(&p.Sku, &p.Price, &p.Cnt, &p.Avatar, &p.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

// func (r *OrderRepository) GetOrdersByUserID(ctx context.Context, userID int64) ([]*protos.Order, error) {
// 	const query = `
//         SELECT
//             user_id,
//             timestamp,
//             product_sku,
//             order_id,
//             status,
//             reason
//         FROM orders
//         WHERE user_id = $1
//     `

// 	rows, err := r.db.Query(ctx, query, userID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to query orders by user_id: %w", err)
// 	}
// 	defer rows.Close()

// 	var orders []*protos.Order
// 	for rows.Next() {
// 		var order protos.Order
// 		if err := rows.Scan(
// 			&order.UserID,
// 			&order.Timestamp,
// 			&order.ProductSKU,
// 			&order.OrderID,
// 			&order.Status,
// 			&order.Reason,
// 		); err != nil {
// 			return nil, fmt.Errorf("failed to scan order: %w", err)
// 		}
// 		orders = append(orders, &order)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, fmt.Errorf("row iteration error: %w", err)
// 	}

// 	return orders, nil
// }

///////////////////////////

// func (r *Repository) CreateOrder(ctx context.Context, order *protos.Order) error {
// 	data, err := proto.Marshal(order)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = r.pool.Exec(ctx, `
// 		INSERT INTO orders (order_id, status, data)
// 		VALUES ($1, $2, $3)
// 		ON CONFLICT (order_id) DO NOTHING
// 	`, order.OrderID, OrderCreated, data)
// 	if err != nil {
// 		return fmt.Errorf("create order: %w", err)
// 	}

// 	return nil
// }

// func (r *Repository) CommitOrder(ctx context.Context, orderID int64) error {
// 	_, err := r.pool.Exec(ctx, `
// 		UPDATE orders SET status = $1 WHERE order_id = $2
// 	`, OrderCommited, orderID)
// 	if err != nil {
// 		return fmt.Errorf("commit order: %w", err)
// 	}
// 	return nil
// }

// func (r *Repository) CancelOrder(ctx context.Context, orderID int64) error {
// 	_, err := r.pool.Exec(ctx, `
// 		DELETE FROM orders WHERE order_id = $1
// 	`, orderID)
// 	if err != nil {
// 		return fmt.Errorf("cancel order: %w", err)
// 	}
// 	return nil
// }
