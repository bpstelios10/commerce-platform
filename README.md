# commerce-platform

A Go learning project: two microservices (`products`, `orders`) in a single Go
workspace, talking to each other over gRPC and each exposing a REST API over
HTTP with `chi`.

- Architecture, tech stack, and the phase-by-phase learning roadmap:
  [TECH.md](TECH.md)
- Full technical review (strengths, findings, recommendations):
  [TECHNICAL_REVIEW.md](TECHNICAL_REVIEW.md)
- Tracked follow-ups from the review: [open-tasks.md](open-tasks.md)

## Project layout

```
services/products/   REST + gRPC server, in-memory repository
services/orders/      REST server, gRPC client of products
shared/               Cross-service logging (zerolog + slog bridge), graceful-shutdown, properties-loading
protos/               .proto definitions (source of truth for generated gRPC code)
```

Each service is its own Go module (`go.mod`) tied together by the root
`go.work`, and follows the same layering:
`http handler → service → repository`, with dependencies injected as
interfaces via constructors (`NewXxx`).

## Prerequisites

- Go (see `go.work` for the exact version)
- `protoc` + Go plugins, only if you need to regenerate gRPC code (see
  [Protobuf / gRPC](#protobuf--grpc) below)

## Ports

| Service  | Protocol | Port |
| -------- | -------- | ---- |
| products | HTTP     | 8082 |
| products | gRPC     | 8092 |
| orders   | HTTP     | 8083 |

`orders` calls `products` over gRPC at `localhost:8092` (hardcoded in
[services/orders/cmd/main.go](services/orders/cmd/main.go) — see
[TECHNICAL_REVIEW.md](TECHNICAL_REVIEW.md) for the plan to externalize this).

## Configuration

| Env var     | Default | Description                                                                                   |
| ----------- | ------- | --------------------------------------------------------------------------------------------- |
| `LOG_LEVEL` | `info`  | zerolog level name (`debug`, `info`, `warn`, `error`, ...), read on startup by both services. |

## Quick start

All common tasks are wired up in the [Makefile](Makefile):

```bash
make run-all            # run both services (products in background, orders in foreground)
make run-products       # run only products (HTTP :8082, gRPC :8092)
make run-orders         # run only orders (HTTP :8083)

ACTIVE_PROFILE=test make run-all      # run both services on the given environment
ACTIVE_PROFILE=test make run-products # run only products on the given environment
ACTIVE_PROFILE=test make run-orders   # run only orders on the given environment

make test-all           # run tests for shared + orders + products
make test-orders        # run tests for orders only
make test-products      # run tests for products only
make test-v             # same as test-all, verbose
make coverage           # run tests with coverage and open HTML reports

make lint               # golangci-lint across all modules
make tidy               # go mod tidy in every module + go work sync
make build              # build orders/products binaries into bin/
make clean              # remove build artifacts and coverage files
```

You can also run/test each module directly with the standard Go tooling,
scoped with `-C`:

```bash
go run  -C services/products ./cmd/...
go test -C services/products ./...

go run  -C services/orders ./cmd/...
go test -C services/orders ./...
```

## Protobuf / gRPC

The `.proto` source of truth lives under `protos/<name>/<version>/`. Generated
Go code is checked into each consuming service under `internal/grpc/`.

Setup (one-time):

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Regenerate code with the Makefile targets instead of calling `protoc`
directly:

```bash
# regenerate the server-side stubs for a service (e.g. products)
make proto-server PROTO=product VERSION=v1

# regenerate the client-side stubs for a consumer (e.g. orders calling products)
make proto-client PROTO=product VERSION=v1 SERVICE=orders
```
