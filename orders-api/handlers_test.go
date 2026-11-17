package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	menu      map[int64]MenuItem
	created   Order
	total     int64
	listed    []Order
	getErr    error
	pingErr   error
	menuErr   error
	createErr error
}

func (store *fakeStore) MenuPrices(_ context.Context, ids []int64) (map[int64]MenuItem, error) {
	if store.menuErr != nil {
		return nil, store.menuErr
	}
	prices := make(map[int64]MenuItem, len(ids))
	for _, id := range ids {
		if item, ok := store.menu[id]; ok {
			prices[id] = item
		}
	}
	return prices, nil
}

func (store *fakeStore) CreateOrder(_ context.Context, items []OrderItem, total int64) (Order, error) {
	if store.createErr != nil {
		return Order{}, store.createErr
	}
	store.total = total
	store.created = Order{ID: 17, Items: items, TotalPKR: total, Status: "new", CreatedAt: time.Unix(1, 0)}
	return store.created, nil
}

func (store *fakeStore) ListOrders(context.Context) ([]Order, error) {
	return store.listed, nil
}

func (store *fakeStore) GetOrder(context.Context, int64) (Order, error) {
	return store.created, store.getErr
}

func (store *fakeStore) Ping(context.Context) error {
	return store.pingErr
}

type fakeQueue struct {
	order   Order
	at      time.Time
	err     error
	pingErr error
}

func (queue *fakeQueue) EnqueueOrder(_ context.Context, order Order, at time.Time) error {
	queue.order = order
	queue.at = at
	return queue.err
}

func (queue *fakeQueue) Ping(context.Context) error {
	return queue.pingErr
}

func TestCreateOrderCalculatesTotalAndEnqueues(t *testing.T) {
	store := &fakeStore{menu: map[int64]MenuItem{
		1: {ID: 1, Name: "Chicken Biryani", PricePKR: 550},
		2: {ID: 2, Name: "Plain Naan", PricePKR: 70},
	}}
	queue := &fakeQueue{}
	api := NewAPI(store, queue, "abc123", "orders-pod")
	api.now = func() time.Time { return time.Unix(100, 0) }
	request := `{"items":[{"menu_item_id":1,"qty":2},{"menu_item_id":2,"qty":1}]}`

	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(request)))

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body)
	}
	if store.total != 1170 {
		t.Fatalf("total = %d, want 1170", store.total)
	}
	if queue.order.ID != 17 || !queue.at.Equal(time.Unix(100, 0)) {
		t.Fatalf("queued order = %#v at %v", queue.order, queue.at)
	}
}

func TestCreateOrderRejectsInvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty items", body: `{"items":[]}`},
		{name: "zero quantity", body: `{"items":[{"menu_item_id":1,"qty":0}]}`},
		{name: "invalid JSON", body: `{"items":`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			api := NewAPI(&fakeStore{}, &fakeQueue{}, "dev", "pod")
			response := httptest.NewRecorder()
			api.Handler().ServeHTTP(response, httptest.NewRequest(
				http.MethodPost, "/api/orders", strings.NewReader(test.body)))
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateOrderRejectsUnknownMenuItem(t *testing.T) {
	api := NewAPI(&fakeStore{menu: map[int64]MenuItem{}}, &fakeQueue{}, "dev", "pod")
	request := `{"items":[{"menu_item_id":90,"qty":1}]}`
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/orders", strings.NewReader(request)))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestInfoHandler(t *testing.T) {
	api := NewAPI(&fakeStore{}, &fakeQueue{}, "commit-42", "orders-7")
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/orders/info", nil))

	var result map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result["version"] != "commit-42" || result["pod"] != "orders-7" {
		t.Fatalf("info = %#v", result)
	}
}

func TestGetMissingOrder(t *testing.T) {
	api := NewAPI(&fakeStore{getErr: ErrOrderNotFound}, &fakeQueue{}, "dev", "pod")
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/orders/1", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestHealthChecksDependencies(t *testing.T) {
	api := NewAPI(&fakeStore{pingErr: errors.New("postgres down")}, &fakeQueue{}, "dev", "pod")
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/orders/healthz", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
