package queue

import "github.com/raufhm/vfc/internal/domain"

type QueueProvider interface {
	Connect() error
	Close() error
	Enqueue(event *domain.Event) error
	Dequeue() (*domain.Event, error)
	GetChannel() <-chan *domain.Event
}
