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
services/products/   REST + gRPC server, PostgreSQL repositories
services/orders/      REST server, PostgreSQL repository, gRPC client of products
shared/               Logging (zerolog), graceful shutdown, YAML config loading, Postgres client + migrations
protos/               .proto definitions (source of truth for generated gRPC code)
```

Each service is its own Go module (`go.mod`) tied together by the root
`go.work`, and follows the same layering:
`http handler → service → repository`, with dependencies injected as
interfaces via constructors (`NewXxx`).

## Prerequisites

- Go (see `go.work` for the exact version)
- Docker with Docker Compose, for PostgreSQL or the full containerized stack
- `protoc` + Go plugins, only if you need to regenerate gRPC code (see
  [Protobuf / gRPC](#protobuf--grpc) below)

## Ports

| Service  | Protocol | Port |
| -------- | -------- | ---- |
| products | HTTP     | 8082 |
| products | gRPC     | 8092 |
| orders   | HTTP     | 8083 |

`orders` calls `products` over gRPC at the address configured under
`products.grpc-client` in [services/orders/config](services/orders/config)
(`localhost:8092` by default).

## Configuration

| Env var     | Default | Description                                                                                   |
| ----------- | ------- | --------------------------------------------------------------------------------------------- |
| `ACTIVE_PROFILE` | unset | Load only `base.yaml`; set `local`, `test`, or `d-compose` to load an overlay. |
| `LOG_LEVEL` | unset | Override YAML `log-level` with a valid zerolog level (`debug`, `info`, `warn`, `error`, ...). |

Per-service settings (ports, database, gRPC client address) live in
`services/<name>/config/*.yaml` and are embedded when the service is compiled.
`base.yaml` is always loaded. With `ACTIVE_PROFILE` unset, the reported profile
is `default` and no overlay is loaded; the base files use local addresses.
Compose explicitly selects `d-compose`, which uses container service names.
`go test` does not automatically select `test.yaml`; that overlay currently only
changes the environment label, not database addresses or credentials.

Log-level precedence is a valid `LOG_LEVEL`, then YAML `log-level`, then `info`.
Currently orders defaults to `debug` via its base YAML; products defaults to
`info`. An unset or invalid override falls through to the next source.

## Error Handling and Logging

Errors have separate internal and external representations:

- Lower layers return errors and add useful operation context with `%w`, for
  example `query orders: %w` in a repository and `get orders: %w` in a service.
- Callers classify errors with `errors.Is` or `errors.As`; wrapping preserves
  the underlying error for classification and logging.
- HTTP and gRPC boundaries map known domain and validation errors to stable
  status codes, error codes, and safe client messages.
- Unexpected errors are logged once at the boundary with the complete wrapped
  error chain, then returned to the client with a generic message. Internal
  database, network, SQL, and service details must not be sent in responses.
- Expected client or domain errors, such as invalid input or not found, do not
  need error-level logs; the request access log records their status. A lower
  layer may log when it performs a meaningful action such as a retry or
  fallback.

This keeps debugging context in private logs without exposing internal
implementation details through the API.

## Database (PostgreSQL)

Both services connect to Postgres on startup, ping it, and run their own
`golang-migrate` migrations (`services/<name>/migrations/*.sql`) before serving
traffic — each service owns its own database (`orders`, `products`) in one
Postgres instance, created by [database/postgres/init.sql](database/postgres/init.sql).

**Build the Go services locally, then run the full stack with Docker Compose.**
`make build` creates Linux ARM64 binaries, suitable for Apple Silicon Macs:

```bash
make build                    # compile both services on your machine
docker compose up --build -d   # package the binaries and start the full stack
docker compose ps             # check service status
docker compose down           # stop it (add -v to also wipe the data volume)
```

On Intel Macs, use `make build GOARCH=amd64` instead. To start only Postgres
and run the Go services directly on the host, use `docker compose up -d postgres`,
then `make run-all`.

Default local credentials are user `commerce`, password `commerce`, as configured
in each service's `base.yaml`. Both services persist data in PostgreSQL.

Compose's `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD` variables
configure the PostgreSQL container only; they do not override the Go services'
embedded database configuration. If changing credentials, update the matching
service YAML and rebuild the binaries/images too. PostgreSQL initialization
variables and [init.sql](database/postgres/init.sql) apply only to a fresh data
volume; they do not change an existing database or user.

**Inspect the database** with `psql` inside the container:

```bash
docker exec -it commerce-postgres sh    # shell inside the container
psql -U commerce                        # then start psql (connects to the `commerce` maintenance db)
```

or in one step from the host:

```bash
docker exec -it commerce-postgres psql -U commerce -d products
```

Useful `psql` commands once connected:

```sql
\l                       -- list databases (expect: commerce, orders, products)
\c products              -- switch to the products database
\dt                      -- list tables in the current database
\d+ products             -- describe table products (and check eg the FK exists)
SELECT * FROM products;  -- inspect rows
\c orders
\dt
SELECT * FROM orders;
\q                       -- quit psql
```

## Quick start

All common tasks are wired up in the [Makefile](Makefile):

```bash
ACTIVE_PROFILE=local make run-all      # products in background, orders in foreground
ACTIVE_PROFILE=local make run-products # HTTP :8082, gRPC :8092
ACTIVE_PROFILE=local make run-orders   # HTTP :8083

ACTIVE_PROFILE=test make run-all      # run both services on the given environment
ACTIVE_PROFILE=test make run-products # run only products on the given environment
ACTIVE_PROFILE=test make run-orders   # run only orders on the given environment

make test-all           # run tests for shared + orders + products
make test-orders        # run tests for orders only
make test-products      # run tests for products only
make test-v             # same as test-all, verbose
make coverage           # run tests with coverage and open HTML reports

make check              # go vet, race-enabled tests, and gofmt check across all modules
make tidy               # go mod tidy in every module + go work sync
make build              # build Linux ARM64 binaries into bin/
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

For host execution, start PostgreSQL first with `docker compose up -d postgres`;
do not also run the Compose service containers on the same ports.

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
