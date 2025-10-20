package repository

import (
	"errors"
	"github.com/raufhm/vfc/internal/domain"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type ProductRepository interface {
	Save(product *domain.Product) error
	Get(productID string) (*domain.Product, error)
	Delete(productID string) error
	Close() error
}
