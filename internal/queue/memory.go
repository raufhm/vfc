package queue

import (
	"github.com/raufhm/vfc/internal/domain"
	"go.uber.org/zap"
)

type InMemoryQueue struct {
	queue  chan *domain.Event
	logger *zap.Logger
}

func NewInMemoryQueue(bufferSize int, logger *zap.Logger) *InMemoryQueue {
	return &InMemoryQueue{
		queue:  make(chan *domain.Event, bufferSize),
		logger: logger,
	}
}

func (q *InMemoryQueue) Connect() error {
	q.logger.Info("InMemoryQueue connected")
	return nil
}

func (q *InMemoryQueue) Close() error {
	close(q.queue)
	q.logger.Info("InMemoryQueue closed")
	return nil
}

func (q *InMemoryQueue) Enqueue(event *domain.Event) error {
	q.queue <- event
	return nil
}

func (q *InMemoryQueue) Dequeue() (*domain.Event, error) {
	event := <-q.queue
	return event, nil
}

func (q *InMemoryQueue) GetChannel() <-chan *domain.Event {
	return q.queue
}
