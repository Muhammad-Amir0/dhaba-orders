package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (store *PostgresStore) MenuPrices(ctx context.Context, ids []int64) (map[int64]MenuItem, error) {
	rows, err := store.pool.Query(ctx,
		"SELECT id, name, price_pkr FROM menu_items WHERE id = ANY($1)", ids)
	if err != nil {
		return nil, fmt.Errorf("query menu items: %w", err)
	}
	defer rows.Close()

	menu := make(map[int64]MenuItem, len(ids))
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.PricePKR); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		menu[item.ID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read menu items: %w", err)
	}
	return menu, nil
}

func (store *PostgresStore) CreateOrder(ctx context.Context, items []OrderItem, total int64) (Order, error) {
	encodedItems, err := json.Marshal(items)
	if err != nil {
		return Order{}, fmt.Errorf("encode order items: %w", err)
	}
	var order Order
	err = store.pool.QueryRow(ctx,
		"INSERT INTO orders (items, total_pkr, status) VALUES ($1, $2, 'new') "+
			"RETURNING id, items, total_pkr, status, created_at",
		encodedItems, total,
	).Scan(&order.ID, &encodedItems, &order.TotalPKR, &order.Status, &order.CreatedAt)
	if err != nil {
		return Order{}, fmt.Errorf("insert order: %w", err)
	}
	if err := json.Unmarshal(encodedItems, &order.Items); err != nil {
		return Order{}, fmt.Errorf("decode saved order items: %w", err)
	}
	return order, nil
}

func (store *PostgresStore) ListOrders(ctx context.Context) ([]Order, error) {
	rows, err := store.pool.Query(ctx,
		"SELECT id, items, total_pkr, status, created_at FROM orders "+
			"ORDER BY created_at DESC, id DESC LIMIT 50")
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]Order, 0, 50)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read orders: %w", err)
	}
	return orders, nil
}

func (store *PostgresStore) GetOrder(ctx context.Context, id int64) (Order, error) {
	var order Order
	var items []byte
	err := store.pool.QueryRow(ctx,
		"SELECT id, items, total_pkr, status, created_at FROM orders WHERE id = $1", id).
		Scan(&order.ID, &items, &order.TotalPKR, &order.Status, &order.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("query order: %w", err)
	}
	if err := json.Unmarshal(items, &order.Items); err != nil {
		return Order{}, fmt.Errorf("decode order items: %w", err)
	}
	return order, nil
}

func scanOrder(row interface{ Scan(...any) error }) (Order, error) {
	var order Order
	var items []byte
	if err := row.Scan(&order.ID, &items, &order.TotalPKR, &order.Status, &order.CreatedAt); err != nil {
		return Order{}, fmt.Errorf("scan order: %w", err)
	}
	if err := json.Unmarshal(items, &order.Items); err != nil {
		return Order{}, fmt.Errorf("decode order items: %w", err)
	}
	return order, nil
}

func (store *PostgresStore) Ping(ctx context.Context) error {
	return store.pool.Ping(ctx)
}
