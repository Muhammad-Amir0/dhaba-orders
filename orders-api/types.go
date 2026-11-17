package main

import (
	"context"
	"errors"
	"time"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderItem struct {
	MenuItemID int64  `json:"menu_item_id"`
	Qty        int64  `json:"qty"`
	Name       string `json:"name"`
	PricePKR   int64  `json:"price_pkr"`
}

type MenuItem struct {
	ID       int64
	Name     string
	PricePKR int64
}

type Order struct {
	ID        int64       `json:"id"`
	Items     []OrderItem `json:"items"`
	TotalPKR  int64       `json:"total_pkr"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

type CreateOrderRequest struct {
	Items []struct {
		MenuItemID int64 `json:"menu_item_id"`
		Qty        int64 `json:"qty"`
	} `json:"items"`
}

type Store interface {
	MenuPrices(context.Context, []int64) (map[int64]MenuItem, error)
	CreateOrder(context.Context, []OrderItem, int64) (Order, error)
	ListOrders(context.Context) ([]Order, error)
	GetOrder(context.Context, int64) (Order, error)
	Ping(context.Context) error
}

type Queue interface {
	EnqueueOrder(context.Context, Order, time.Time) error
	Ping(context.Context) error
}
