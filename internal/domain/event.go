package domain

import "time"

type Event struct {
	ProductID string    `json:"product_id"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	Timestamp time.Time `json:"timestamp"`
}

func NewEvent(productID string, price float64, stock int) *Event {
	return &Event{
		ProductID: productID,
		Price:     price,
		Stock:     stock,
		Timestamp: time.Now(),
	}
}

func (e *Event) ToProduct() *Product {
	return &Product{
		ProductID: e.ProductID,
		Price:     e.Price,
		Stock:     e.Stock,
		UpdatedAt: e.Timestamp,
	}
}
