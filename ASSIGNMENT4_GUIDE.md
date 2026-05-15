# Assignment 4 - Performance Optimization & External Integrations Guide

This document is your guide for the Week 9 Defense for Assignment 4.

## 1. Caching with Redis (Cache-Aside & Invalidation)
**Location:** `order-service/internal/usecase/order_usecase.go`
*   **Cache-Aside Read Path:** In `GetOrder(id string)`, the code first attempts to fetch the order from Redis (`uc.cache.Get(id)`). If there is a cache miss, it falls back to PostgreSQL, and then immediately writes the result into Redis (`uc.cache.Set(order)`) with a TTL of 5 minutes.
*   **Atomic Invalidation:** When an order is updated (for instance, after a successful payment changes it from `Pending` to `Paid`, or when it is explicitly Cancelled), the usecase calls `uc.cache.Invalidate(order.ID)`. This performs a Redis `DEL` command to completely wipe the cached record, strictly preventing the system from ever serving a stale "Pending" state.

## 2. External Provider Adapter (Notification Service)
**Location:** `notification-service/internal/infrastructure/providers.go`
*   **Adapter Pattern:** The core logic now uses the `domain.NotificationProvider` interface, allowing us to swap out email vendors without touching the business logic.
*   **Simulated Provider:** I implemented `SimulatedProvider` which introduces realistic network latency (`time.Sleep`) and intentionally triggers random failures (a 20% simulated crash rate) to actually put our background worker's retry logic to the test.
*   **Configuration:** The service boots up reading the `PROVIDER_MODE` environment variable (set to `SIMULATED` in `docker-compose.yml`), resolving the provider choice at startup.

## 3. Reliable Background Worker with Retries & Idempotency
**Location:** `notification-service/internal/usecase/notification_worker.go`
*   **Redis Idempotency:** Instead of using an in-memory map, the idempotency check is now correctly managed across a distributed system using Redis. It stores the `message_id` with a 24-hour TTL when a notification successfully processes. If the same RabbitMQ message is redelivered, the worker sees it in Redis and gracefully drops the duplicate.
*   **Exponential Backoff Retries:** If the simulated provider throws an error, the worker executes a resilient `for` loop, retrying up to 3 times. Crucially, the wait time doubles after every failure (`2s -> 4s -> 8s`), giving the external API time to recover before hammering it again.

## 4. Bonus (+10%): Redis Rate Limiter
**Location:** `order-service/internal/transport/http/middleware.go`
*   **Middleware:** I built a custom Gin Middleware that protects our Order REST API. It identifies clients by their IP address and creates a Redis key (`ratelimit:IP`).
*   **Counter Increment:** It utilizes Redis `INCR` to keep a highly performant distributed tally of requests. Once it increments past our limit (10 requests), it instantly rejects the traffic and returns an HTTP `429 Too Many Requests` error message. It correctly resets the counter utilizing Redis `EXPIRE` mapped to a 1-minute sliding window.

## 5. Docker Infrastructure
**Location:** `docker-compose.yml`
*   A `redis:7-alpine` container was added to the orchestration. Both the Order Service and Notification Service connect to it seamlessly via the internal Docker network.
