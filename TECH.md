# Go Backend Learning Project

## Goal

Learn production-grade Go backend development from the perspective of an experienced Java engineer.

The objective is not only to learn Go syntax, but also the common ecosystem, tooling, architecture patterns, and operational practices expected from a Go backend engineer.

---

# Final Architecture

Two microservices:

1. Product Service
2. Order Service

Each service owns its own database schema.

Communication will happen using:

- REST
- gRPC
- Kafka Events

---

# Product Service

Responsibilities:

- Create products
- Retrieve products
- Manage inventory

Exposes:

- REST API
- gRPC API

Database:

- PostgreSQL

---

# Order Service

Responsibilities:

- Create orders
- Validate products
- Manage order lifecycle

Exposes:

- gRPC API only

Database:

- PostgreSQL

Consumes:

- Product Service

Publishes:

- OrderCreated events

---

# Learning Phases

---

## Phase 1 - Go Fundamentals

### Goal

Learn the language itself.

### Concepts

- Modules
- Packages
- Structs
- Methods
- Interfaces
- Error handling
- Pointers

### Deliverable

Simple CLI application.

### Learn

Difference between Java OOP and Go composition.

---

## Phase 2 - Product Service HTTP API

### Goal

Build first real service.

### Technologies

- net/http
- Chi Router

### Concepts

- Routing
- Middleware
- JSON serialization
- Request validation

### Deliverable

Endpoints:

POST /products

GET /products/{id}

### Learn

How Go handles HTTP without Spring.

---

## Phase 3 - PostgreSQL Integration

### Goal

Persist products.

### Technologies

- PostgreSQL
- pgx
- goose migrations

### Concepts

- Repository pattern
- Transactions
- Connection pooling

### Deliverable

Persist products in database.

### Learn

Idiomatic database access in Go.

---

## Phase 4 - Configuration

### Goal

Externalize configuration.

### Technologies

- Environment Variables

### Concepts

- Startup configuration
- Dependency injection

### Deliverable

Database configuration loaded from env vars.

---

## Phase 5 - Logging

### Goal

Introduce observability.

### Technologies

- slog

### Concepts

- Structured logging
- Correlation IDs

### Deliverable

All requests logged.

---

## Phase 6 - Testing

### Goal

Learn Go testing patterns.

### Technologies

- testing
- testify

### Concepts

- Unit tests
- Table-driven tests
- Mocking

### Deliverable

Service layer fully tested.

### Learn

The Go testing style.

---

## Phase 7 - Docker

### Goal

Containerize services.

### Technologies

- Docker
- Docker Compose

### Deliverable

Product Service + PostgreSQL running locally.

---

## Phase 8 - gRPC

### Goal

Learn internal service communication.

### Technologies

- Protocol Buffers
- grpc-go

### Deliverable

Product Service exposes:

GetProduct()

CreateProduct()

### Learn

Why Go teams heavily use gRPC.

---

## Phase 9 - Order Service

### Goal

Build second microservice.

### Technologies

- grpc-go

### Deliverable

CreateOrder()

### Workflow

1. Receive gRPC request
2. Call Product Service
3. Validate inventory
4. Create order

### Learn

Service-to-service communication.

---

## Phase 10 - REST Client

### Goal

Compare REST and gRPC.

### Deliverable

Order Service can call Product Service using:

- REST
- gRPC

switchable by configuration.

### Learn

Tradeoffs between both approaches.

---

## Phase 11 - Context Propagation

### Goal

Understand Go request lifecycle.

### Concepts

- context.Context
- cancellation
- timeouts

### Deliverable

Request deadlines propagated between services.

### Learn

One of the most important Go concepts.

---

## Phase 12 - Kafka

### Goal

Event-driven communication.

### Technologies

- Kafka

### Deliverable

OrderCreated event published.

### Learn

Async communication patterns.

---

## Phase 13 - Kafka Consumer

### Goal

Consume events.

### Deliverable

Product Service consumes OrderCreated.

Inventory gets updated.

### Learn

Consumer groups
Retries
Idempotency

---

## Phase 14 - Concurrency

### Goal

Learn Go concurrency.

### Concepts

- Goroutines
- Channels
- WaitGroups
- Worker Pools

### Deliverable

Background inventory processor.

### Learn

The Go concurrency model.

---

## Phase 15 - Metrics

### Goal

Introduce monitoring.

### Technologies

- Prometheus

### Deliverable

/metrics endpoint

Metrics:

- request count
- latency
- error count

---

## Phase 16 - Distributed Tracing

### Goal

Trace requests across services.

### Technologies

- OpenTelemetry

### Deliverable

Single trace spanning:

Order Service
→ Product Service

### Learn

Production observability.

---

## Phase 17 - Integration Testing

### Goal

Test entire workflows.

### Technologies

- testcontainers-go

### Deliverable

Automated integration tests using PostgreSQL.

---

## Phase 18 - Production Hardening

### Goal

Make services production-ready.

### Concepts

- Graceful shutdown
- Health checks
- Readiness checks
- Retry policies

### Deliverable

Production-grade services.

---

# Final Skills Achieved

By completing all phases you should be comfortable with:

- Go language fundamentals
- REST APIs
- gRPC
- PostgreSQL
- Kafka
- Concurrency
- Context propagation
- Logging
- Metrics
- Tracing
- Docker
- Testing
- Microservice communication

This represents the core skill set expected from a backend Go engineer in most modern companies.
