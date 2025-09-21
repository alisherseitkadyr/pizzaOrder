package web

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-system/services/tracking-service/domain/service"
	"strings"
)

type WebHandler struct {
	TrackingService *service.TrackingService
}

func NewWebHandler(trackingService *service.TrackingService) *WebHandler {
	return &WebHandler{TrackingService: trackingService}
}

func (h *WebHandler) GetOrderStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	// Extract order number from URL path
	orderNumber := extractOrderNumber(r.URL.Path, "/orders/", "/status")
	if orderNumber == "" {
		log.Printf("Invalid order number in path: %s", r.URL.Path)
		http.Error(w, "Invalid order number", http.StatusBadRequest)
		return
	}

	status, err := h.TrackingService.GetOrderStatus(r.Context(), orderNumber)
	if err != nil {
		log.Printf("Error getting order status: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *WebHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	// Extract order number from URL path
	orderNumber := extractOrderNumber(r.URL.Path, "/orders/", "/history")
	if orderNumber == "" {
		log.Printf("Invalid order number in path: %s", r.URL.Path)
		http.Error(w, "Invalid order number", http.StatusBadRequest)
		return
	}

	history, err := h.TrackingService.GetOrderHistory(r.Context(), orderNumber)
	if err != nil {
		log.Printf("Error getting order history: %v", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *WebHandler) GetWorkersStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	workers, err := h.TrackingService.GetWorkersStatus(r.Context())
	if err != nil {
		log.Printf("Error getting workers status: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

// Helper function to extract order number from URL path
func extractOrderNumber(path, prefix, suffix string) string {
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return ""
	}

	// Remove prefix and suffix
	orderNumber := strings.TrimPrefix(path, prefix)
	orderNumber = strings.TrimSuffix(orderNumber, suffix)

	// Ensure we have a valid order number
	if orderNumber == "" {
		return ""
	}

	return orderNumber
}
