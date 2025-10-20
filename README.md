# Product Update Service

## Setup Instructions

### Prerequisites
- Go installed on your local machine

### Running the Service

```bash
# Clone and navigate to the project
git clone 
cd vfc

# Sync dependencies
go mod tidy

# Run the service
go run ./cmd/server/main.go
```

The service starts on `http://localhost:8080`. Configuration can be customized in `.env` file (port, worker count, queue buffer size). the guide is under .env.example

### Testing

```bash
# Run all tests
make test

# Test the API
curl http://localhost:8080/health

curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"product_id":"abc123","price":49.99,"stock":100}'

curl http://localhost:8080/products/abc123
```

## Design Choices

### Clean Architecture Approach

The service follows clean architecture principles with clear separation of concerns. Each layer has a specific responsibility and communicates through well-defined interfaces. This design makes it easy to swap implementations without changing business logic.

**The layers:**
- **Domain** - Core business entities (Product, Event) with no external dependencies
- **Repository** - Data access abstraction with thread-safe in-memory implementation
- **Service** - Orchestrates business logic between repository and queue
- **Handler** - HTTP layer using Gorilla Mux for routing
- **Worker** - Background processing with configurable worker pool

This structure is intentional. When we need to add PostgreSQL or RabbitMQ in production, we simply create new implementations of the existing interfaces without touching business logic.

### Concurrency Safety

Thread safety is handled at the repository level using `sync.RWMutex`. This allows multiple concurrent reads while ensuring exclusive access for writes. The worker pool uses Go channels for thread-safe queuing, and graceful shutdown is managed through context cancellation.
The key design decision here was to make concurrency safety explicit at the infrastructure layer (repository, queue) rather than scattering locks throughout the codebase.

### Technology Choices

**Gorilla Mux** was chosen for HTTP routing because it provides clean path parameter extraction and middleware support without the complexity of larger frameworks.
**Viper** handles configuration from `.env` files, making it easy to adjust settings per environment without code changes.
**Zap** provides structured logging with good performance. In production, structured logs make it much easier to search, filter, and analyze issues.

### How It Works

When an event arrives via POST `/events`, it's immediately enqueued and returns `202 Accepted`. The worker pool runs in the background, pulling events from the queue and updating the repository. This asynchronous approach means the API remains fast even under load, and we can scale workers independently of API handlers.
Later events for the same product override earlier ones, which is the expected behavior for product updates.

## Testing

The service includes three types of tests:

**API Tests** - Verify endpoints work correctly, validate input, and return proper status codes
**Concurrency Tests** - Ensure the repository handles concurrent reads and writes safely without race conditions
**Worker Tests** - Confirm the worker pool processes events correctly and shuts down gracefully

All tests pass with the race detector enabled, confirming there are no concurrency issues.

## Production Considerations

### Current Architecture Benefits

The clean architecture we've implemented already provides a solid foundation for production. Because we use interfaces throughout, migrating to production-grade tools is straightforward - we implement the interface for the new technology and swap it in.

### Message Queue (RabbitMQ)

The in-memory queue works fine for this demo, but production needs durability. **RabbitMQ** would give us:

**Message Persistence** - Events survive service restarts, so nothing gets lost during deployments or crashes

**Acknowledgment System** - Messages aren't removed from the queue until successfully processed

**Dead Letter Queues** - Failed messages go to a separate queue for investigation rather than disappearing

**Multiple Consumers** - If we deploy multiple service instances, RabbitMQ distributes work across all of them automatically

Since we already have the `QueueProvider` interface, adding RabbitMQ is just a matter of implementing `Connect`, `Enqueue`, `GetChannel`, and `Close` methods for RabbitMQ client library.

### Database Persistence (PostgreSQL with Bun)

Right now products live in memory, which means they're lost on restart. **PostgreSQL** provides:

**Data Durability** - Products persist across restarts

**ACID Transactions** - Ensures data integrity even under concurrent updates

**Query Capabilities** - Can search, filter, and aggregate products as needed

**Scalability** - Read replicas can handle GET requests while the primary handles writes

Bun ORM (from Uptrace) makes this integration clean since it provides a type-safe query builder. The repository interface already matches what we need, just implement `Save`, `Get`, `Delete` with database calls instead of map operations.

### Caching (Redis)

Adding **Redis** in front of the database reduces load and improves response times:

**Cache Reads** - GET requests hit Redis first, only querying the database on cache miss

**TTL Management** - Set appropriate expiration times based on how often prices change

**Cache Invalidation** - When a product updates, invalidate its cache entry so the next read gets fresh data

The pattern is straightforward: check cache, if miss then query database and populate cache, if hit then return cached value.

### Scaling for High Throughput

**Performance Tuning** - The `.env` file lets you tune performance without code changes: try increase `WORKER_COUNT` to process more events simultaneously, increase `QUEUE_BUFFER_SIZE` to handle traffic spike, and adjust timeouts based on your infrastructure.

**Rate Limiting** - In production, add rate limiting per API client to prevent abuse and ensure fair resource usage. 

### Error Handling

**Retry Logic** - Transient failures (network blips, temporary database unavailability) should be retried with exponential backoff. Start with a short delay, double it on each retry, up to a maximum. After a configured number of attempts, move the event to a dead letter queue for manual investigation.

**Idempotency** - Give each event a unique ID and track processed IDs. This prevents duplicate processing if the same event arrives multiple times (which can happen with message queue retries).

**Circuit Breaker** - If the database or another dependency becomes unavailable, a circuit breaker fails fast rather than letting requests pile up. After the circuit "opens", it periodically checks if the dependency has recovered before resuming normal operation.

## Troubleshooting Strategies

### Data Consistency Problems

**If you see stale or incorrect product data:**

First, check for race conditions by running tests with the race detector: `go test -race ./...`. This catches most concurrency bugs immediately. Next, verify the repository is properly locking. Every write operation should acquire the mutex, every read should acquire at least a read lock. Check that defers are releasing locks correctly. Look at the logs to see if events are processing in unexpected order. Search for "Processing event" entries for the affected product and verify the sequence makes sense. Finally, make sure the worker pool started. Look for "Worker started" log entries. If they're missing, the workers never began consuming events.

### Products Not Updating Despite Events Being Received

**Symptom:** API accepts events (202 response) but products don't change.

This is typically a pipeline break somewhere between the API and the repository. Debug it step by step:

**Step 1 - Check Enqueuing** - Search logs for "Event enqueued" for your product. If it's there, the event made it to the queue. If not, either the handler isn't being called or the queue buffer is full.

**Step 2 - Check Workers Started** - Look for "Worker started" entries in the logs. You should see one per worker (default is 3). If missing, `pool.Start()` wasn't called in main.go.

**Step 3 - Check Processing** - Search for "Processing event" for your product. If present, workers are consuming events. If missing, workers might have panicked or there's a deadlock.

**Step 4 - Check Repository Updates** - Look for "Product updated successfully" for your product. If missing, the repository save is failing for some reason.

**Common fixes:** Increase `QUEUE_BUFFER_SIZE` if the queue fills up during traffic spikes, ensure proper startup order (initialize queue, start workers, then start HTTP server), check shutdown order (stop workers, shut down server, then close queue), and look for panic stack traces that would kill workers.

**Quick test:**
```bash
curl -X POST http://localhost:8080/events \
  -d '{"product_id":"test123","price":99.99,"stock":50}'

sleep 1

curl http://localhost:8080/products/test123
```

If the product returns with the correct price and stock, the pipeline is working. If not, follow the debugging steps above.

### Using Workflow Orchestration for Visibility

In production systems with complex event processing, traditional logging can make debugging painful. You're searching through logs trying to piece together what happened across multiple services and steps.

**Temporal.io or Windmill.dev** solve this visibility problem:

Instead of stitching together logs, you see a visual representation of your workflow with the exact step that failed highlighted. For a product update that involves validating the event, updating the database, invalidating cache, and publishing a notification, you immediately see which step broke.
These tools provide execution history showing every workflow run, retry and timeout handling built-in, state persistence across failures, and the ability to replay failed workflows after fixing bugs. This is particularly valuable during troubleshooting - you can see the entire workflow execution timeline, inspect the input/output of each step, and identify the failure point in seconds.
Think of it as upgrading from grepping logs to having a debugger with time-travel capabilities for your distributed system.
