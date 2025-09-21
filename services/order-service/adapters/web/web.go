package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"restaurant-system/services/order-service/domain/models"
	"restaurant-system/services/order-service/domain/service"
	"restaurant-system/services/order-service/utils/logger"
	"strings"
	"time"
)

type WebHandler struct {
	OrderService *service.OrderService
	Logger       *logger.Logger
}

func NewWebHandler(orderService *service.OrderService, serviceName string) *WebHandler {
	return &WebHandler{
		OrderService: orderService,
		Logger:       logger.New(serviceName),
	}
}

func (h *WebHandler) HandleOrder(w http.ResponseWriter, r *http.Request) {
	// Generate request ID for tracing
	requestID := generateRequestID()

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Log request
	h.Logger.Info("request_received", "Received order creation request", requestID)

	// Validate HTTP method
	if r.Method != http.MethodPost {
		h.Logger.Error("invalid_method", "Invalid HTTP method", requestID,
			httpErrorf(http.StatusMethodNotAllowed, "Method not allowed"))
		http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var request models.OrderCreatedRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.Logger.Error("invalid_json", "Invalid JSON format", requestID, err)
		http.Error(w, `{"error": "Invalid JSON format"}`, http.StatusBadRequest)
		return
	}

	// Log order received
	h.Logger.Debug("order_received", "Order validation started", requestID)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Create order using service
	order, err := h.OrderService.CreateOrder(
		ctx,
		request.CustomerName,
		request.OrderType,
		request.Items,
		request.TableNumber,
		request.DeliveryAddress,
	)
	if err != nil {
		h.Logger.Error("order_creation_failed", "Failed to create order", requestID, err)

		// Check error type and return appropriate status code
		if strings.Contains(err.Error(), "validation") {
			sendJSONError(w, http.StatusBadRequest, err.Error())
		} else if strings.Contains(err.Error(), "RabbitMQ") {
			// RabbitMQ errors are still considered successful order creation
			// but we should log them as warnings
			h.Logger.Error("rabbitmq_publish_failed", "Order saved but RabbitMQ publish failed", requestID, err)
			// Continue with success response since order was saved
		} else {
			sendJSONError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	// Log successful order creation
	h.Logger.Debug("order_created", "Order created successfully", requestID)

	// Respond with the created order according to TZ specification
	response := models.CreateOrderResponse{
		OrderNumber: order.OrderNumber,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Error("response_encode_failed", "Failed to encode response", requestID, err)
	}
}

// Helper function to send JSON errors
func sendJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Helper function to generate request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Custom error type for HTTP errors
type httpError struct {
	statusCode int
	message    string
}

func (e *httpError) Error() string {
	return e.message
}

func httpErrorf(statusCode int, format string, args ...interface{}) error {
	return &httpError{
		statusCode: statusCode,
		message:    fmt.Sprintf(format, args...),
	}
}
