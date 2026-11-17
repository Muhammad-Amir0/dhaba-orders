package main

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

type API struct {
	store   Store
	queue   Queue
	version string
	pod     string
	now     func() time.Time
}

func NewAPI(store Store, queue Queue, version, pod string) *API {
	return &API{store: store, queue: queue, version: version, pod: pod, now: time.Now}
}

func (api *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/orders", api.createOrder)
	mux.HandleFunc("GET /api/orders", api.listOrders)
	mux.HandleFunc("GET /api/orders/{id}", api.getOrder)
	mux.HandleFunc("GET /api/orders/info", api.info)
	mux.HandleFunc("GET /api/orders/healthz", api.healthz)
	mux.HandleFunc("GET /api/orders/livez", api.livez)
	return mux
}

func (api *API) createOrder(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var request CreateOrderRequest
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(request.Items) == 0 || len(request.Items) > 100 {
		writeError(w, http.StatusBadRequest, "items must contain between 1 and 100 entries")
		return
	}

	ids := make([]int64, 0, len(request.Items))
	for _, item := range request.Items {
		if item.MenuItemID < 1 || item.Qty < 1 || item.Qty > 100 {
			writeError(w, http.StatusBadRequest, "menu_item_id and qty must be positive")
			return
		}
		ids = append(ids, item.MenuItemID)
	}

	menu, err := api.store.MenuPrices(r.Context(), ids)
	if err != nil {
		log.Printf("look up menu prices: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create order")
		return
	}

	items := make([]OrderItem, 0, len(request.Items))
	var total int64
	for _, requested := range request.Items {
		menuItem, exists := menu[requested.MenuItemID]
		if !exists {
			writeError(w, http.StatusBadRequest, "unknown menu_item_id")
			return
		}
		if menuItem.PricePKR > math.MaxInt64/requested.Qty {
			writeError(w, http.StatusBadRequest, "order total is too large")
			return
		}
		lineTotal := menuItem.PricePKR * requested.Qty
		if total > math.MaxInt64-lineTotal {
			writeError(w, http.StatusBadRequest, "order total is too large")
			return
		}
		total += lineTotal
		items = append(items, OrderItem{
			MenuItemID: requested.MenuItemID,
			Qty:        requested.Qty,
			Name:       menuItem.Name,
			PricePKR:   menuItem.PricePKR,
		})
	}

	order, err := api.store.CreateOrder(r.Context(), items, total)
	if err != nil {
		log.Printf("save order: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create order")
		return
	}
	if err := api.queue.EnqueueOrder(r.Context(), order, api.now()); err != nil {
		log.Printf("enqueue order %d: %v", order.ID, err)
		writeError(w, http.StatusInternalServerError, "order saved but could not be queued")
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (api *API) listOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := api.store.ListOrders(r.Context())
	if err != nil {
		log.Printf("list orders: %v", err)
		writeError(w, http.StatusInternalServerError, "could not list orders")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (api *API) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	order, err := api.store.GetOrder(r.Context(), id)
	if errors.Is(err, ErrOrderNotFound) {
		writeError(w, http.StatusNotFound, "order not found")
		return
	}
	if err != nil {
		log.Printf("get order %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "could not get order")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (api *API) info(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": api.version,
		"pod":     api.pod,
	})
}

func (api *API) healthz(w http.ResponseWriter, r *http.Request) {
	if err := api.store.Ping(r.Context()); err != nil {
		log.Printf("postgres readiness check: %v", err)
		writeError(w, http.StatusServiceUnavailable, "dependency unavailable")
		return
	}
	if err := api.queue.Ping(r.Context()); err != nil {
		log.Printf("redis readiness check: %v", err)
		writeError(w, http.StatusServiceUnavailable, "dependency unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": api.version,
		"pod":     api.pod,
	})
}

func (api *API) livez(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func minuteBucket(now time.Time) string {
	return now.UTC().Format("2006-01-02T15:04Z")
}
