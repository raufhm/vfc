package repository

import (
	"github.com/raufhm/vfc/internal/domain"
	"sync"
)

type InMemoryRepository struct {
	mu       sync.RWMutex
	products map[string]*domain.Product
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		products: make(map[string]*domain.Product),
	}
}

func (r *InMemoryRepository) Save(product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[product.ProductID] = product
	return nil
}

func (r *InMemoryRepository) Get(productID string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[productID]
	if !exists {
		return nil, ErrProductNotFound
	}

	return &domain.Product{
		ProductID: product.ProductID,
		Price:     product.Price,
		Stock:     product.Stock,
		UpdatedAt: product.UpdatedAt,
	}, nil
}

func (r *InMemoryRepository) Delete(productID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.products, productID)
	return nil
}

func (r *InMemoryRepository) Close() error {
	return nil
}

func (r *InMemoryRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.products)
}
