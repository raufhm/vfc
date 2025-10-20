package service

import (
	"github.com/raufhm/vfc/internal/domain"
	"github.com/raufhm/vfc/internal/queue"
	"github.com/raufhm/vfc/internal/repository"
)

// ProductService handles business logic for products
type ProductService struct {
	repo  repository.ProductRepository
	queue queue.QueueProvider
}

// NewProductService creates a new product service
func NewProductService(repo repository.ProductRepository, queue queue.QueueProvider) *ProductService {
	return &ProductService{
		repo:  repo,
		queue: queue,
	}
}

// EnqueueProductUpdate enqueues a product update event
func (s *ProductService) EnqueueProductUpdate(event *domain.Event) error {
	return s.queue.Enqueue(event)
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(productID string) (*domain.Product, error) {
	return s.repo.Get(productID)
}
