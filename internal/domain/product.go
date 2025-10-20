package domain

import "time"

type Product struct {
	ProductID string    `json:"product_id"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewProduct(productID string, price float64, stock int) *Product {
	return &Product{
		ProductID: productID,
		Price:     price,
		Stock:     stock,
		UpdatedAt: time.Now(),
	}
}
