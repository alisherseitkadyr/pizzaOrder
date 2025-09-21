package web

import "net/http"

func NewRouter(handler *WebHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /orders/{order_number}/status", handler.GetOrderStatus)
	mux.HandleFunc("GET /orders/{order_number}/history", handler.GetOrderHistory)
	mux.HandleFunc("GET /workers/status", handler.GetWorkersStatus)

	return mux
}
