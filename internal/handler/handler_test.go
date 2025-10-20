package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/raufhm/vfc/internal/domain"
	"github.com/raufhm/vfc/internal/queue"
	"github.com/raufhm/vfc/internal/repository"
	"github.com/raufhm/vfc/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTest() (*ProductHandler, *repository.InMemoryRepository, *queue.InMemoryQueue) {
	logger, _ := zap.NewDevelopment()
	repo := repository.NewInMemoryRepository()
	q := queue.NewInMemoryQueue(10, logger)
	svc := service.NewProductService(repo, q)
	handler := NewProductHandler(svc, logger)
	return handler, repo, q
}

func TestCreateEvent_Success(t *testing.T) {
	handler, _, _ := setupTest()

	reqBody := EventRequest{
		ProductID: "test123",
		Price:     49.99,
		Stock:     100,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.CreateEvent(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)
}

func TestCreateEvent_InvalidJSON(t *testing.T) {
	handler, _, _ := setupTest()

	req := httptest.NewRequest("POST", "/events", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.CreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid request body", errResp.Error)
}

func TestCreateEvent_MissingProductID(t *testing.T) {
	handler, _, _ := setupTest()

	reqBody := EventRequest{
		Price: 49.99,
		Stock: 100,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.CreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "product_id is required", errResp.Error)
}

func TestCreateEvent_NegativePrice(t *testing.T) {
	handler, _, _ := setupTest()

	reqBody := EventRequest{
		ProductID: "test123",
		Price:     -10.0,
		Stock:     100,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.CreateEvent(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "price must be non-negative", errResp.Error)
}

func TestGetProduct_Success(t *testing.T) {
	handler, repo, _ := setupTest()

	product := domain.NewProduct("test123", 49.99, 100)
	err := repo.Save(product)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/products/test123", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test123"})

	rr := httptest.NewRecorder()

	handler.GetProduct(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp domain.Product
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "test123", resp.ProductID)
	assert.Equal(t, 49.99, resp.Price)
	assert.Equal(t, 100, resp.Stock)
}

func TestGetProduct_NotFound(t *testing.T) {
	handler, _, _ := setupTest()

	req := httptest.NewRequest("GET", "/products/nonexistent", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})

	rr := httptest.NewRecorder()

	handler.GetProduct(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var errResp ErrorResponse
	err := json.NewDecoder(rr.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Product not found", errResp.Error)
}

func TestHealthCheck(t *testing.T) {
	handler, _, _ := setupTest()

	req := httptest.NewRequest("GET", "/health", nil)

	rr := httptest.NewRecorder()

	handler.HealthCheck(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
}
