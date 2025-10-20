package worker

import (
	"context"
	"sync"

	"github.com/raufhm/vfc/internal/domain"
	"github.com/raufhm/vfc/internal/queue"
	"github.com/raufhm/vfc/internal/repository"
	"go.uber.org/zap"
)

type Pool struct {
	workerCount int
	queue       queue.QueueProvider
	repo        repository.ProductRepository
	logger      *zap.Logger
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewPool(workerCount int, queue queue.QueueProvider, repo repository.ProductRepository, logger *zap.Logger) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workerCount: workerCount,
		queue:       queue,
		repo:        repo,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (p *Pool) Start() {
	p.logger.Info("Starting worker pool", zap.Int("worker_count", p.workerCount))

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i + 1)
	}
}

func (p *Pool) Stop() {
	p.logger.Info("Stopping worker pool")
	p.cancel()
	p.wg.Wait()
	p.logger.Info("Worker pool stopped")
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	p.logger.Info("Worker started", zap.Int("worker_id", id))

	eventChan := p.queue.GetChannel()

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info("Worker stopping", zap.Int("worker_id", id))
			return
		case event, ok := <-eventChan:
			if !ok {
				p.logger.Info("Queue closed", zap.Int("worker_id", id))
				return
			}

			if event == nil {
				continue
			}

			p.processEvent(id, event)
		}
	}
}

func (p *Pool) processEvent(workerID int, event *domain.Event) {
	p.logger.Info("Processing event",
		zap.Int("worker_id", workerID),
		zap.String("product_id", event.ProductID))

	product := event.ToProduct()

	if err := p.repo.Save(product); err != nil {
		p.logger.Error("Failed to save product",
			zap.Int("worker_id", workerID),
			zap.String("product_id", product.ProductID),
			zap.Error(err))
		return
	}

	p.logger.Info("Product updated successfully",
		zap.Int("worker_id", workerID),
		zap.String("product_id", product.ProductID),
		zap.Float64("price", product.Price),
		zap.Int("stock", product.Stock))
}
