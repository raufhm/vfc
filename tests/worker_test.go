package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/raufhm/vfc/internal/domain"
	"github.com/raufhm/vfc/internal/queue"
	"github.com/raufhm/vfc/internal/repository"
	"github.com/raufhm/vfc/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWorkerPoolProcessing(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := repository.NewInMemoryRepository()
	q := queue.NewInMemoryQueue(100, logger)
	pool := worker.NewPool(3, q, repo, logger)

	pool.Start()

	for i := 0; i < 10; i++ {
		event := domain.NewEvent("product-1", float64(i)*10, i*100)
		err := q.Enqueue(event)
		require.NoError(t, err)
	}

	time.Sleep(500 * time.Millisecond)
	pool.Stop()

	product, err := repo.Get("product-1")
	require.NoError(t, err)
	assert.Equal(t, "product-1", product.ProductID)
}

func TestMultipleProducts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := repository.NewInMemoryRepository()
	q := queue.NewInMemoryQueue(100, logger)
	pool := worker.NewPool(5, q, repo, logger)

	pool.Start()

	for i := 0; i < 20; i++ {
		productID := fmt.Sprintf("product-%d", i)
		event := domain.NewEvent(productID, float64(i)*10, i*100)
		err := q.Enqueue(event)
		require.NoError(t, err)
	}

	time.Sleep(1 * time.Second)
	pool.Stop()

	for i := 0; i < 20; i++ {
		productID := fmt.Sprintf("product-%d", i)
		product, err := repo.Get(productID)
		require.NoError(t, err)
		assert.Equal(t, productID, product.ProductID)
		assert.Equal(t, float64(i)*10, product.Price)
		assert.Equal(t, i*100, product.Stock)
	}
}

func TestGracefulShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	repo := repository.NewInMemoryRepository()
	q := queue.NewInMemoryQueue(10, logger)
	pool := worker.NewPool(3, q, repo, logger)

	pool.Start()

	for i := 0; i < 5; i++ {
		event := domain.NewEvent("test", float64(i), i)
		err := q.Enqueue(event)
		require.NoError(t, err)
	}

	time.Sleep(100 * time.Millisecond)

	done := make(chan bool)
	go func() {
		pool.Stop()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Worker pool did not shut down within timeout")
	}
}
