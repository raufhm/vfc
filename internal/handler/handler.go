package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/raufhm/vfc/internal/domain"
	"github.com/raufhm/vfc/internal/repository"
	"github.com/raufhm/vfc/internal/service"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service *service.ProductService
	logger  *zap.Logger
}

func NewProductHandler(service *service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

type EventRequest struct {
	ProductID string  `json:"product_id"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *ProductHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req EventRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProductID == "" {
		h.sendError(w, "product_id is required", http.StatusBadRequest)
		return
	}

	if req.Price < 0 {
		h.sendError(w, "price must be non-negative", http.StatusBadRequest)
		return
	}

	if req.Stock < 0 {
		h.sendError(w, "stock must be non-negative", http.StatusBadRequest)
		return
	}

	event := domain.NewEvent(req.ProductID, req.Price, req.Stock)

	if err := h.service.EnqueueProductUpdate(event); err != nil {
		h.logger.Error("Failed to enqueue event", zap.Error(err))
		h.sendError(w, "Failed to enqueue event", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Event enqueued", zap.String("product_id", req.ProductID))

	w.WriteHeader(http.StatusAccepted)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID := vars["id"]

	if productID == "" {
		h.sendError(w, "product_id is required", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProduct(productID)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			h.sendError(w, "Product not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get product", zap.Error(err))
		h.sendError(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, product, http.StatusOK)
}

func (h *ProductHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.sendJSON(w, map[string]string{"status": "healthy"}, http.StatusOK)
}

func (h *ProductHandler) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON", zap.Error(err))
	}
}

func (h *ProductHandler) sendError(w http.ResponseWriter, message string, status int) {
	h.sendJSON(w, ErrorResponse{Error: message}, status)
}
