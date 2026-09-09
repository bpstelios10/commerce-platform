# Technical Review — commerce-platform

Reviewed: 2026-08-31. Scope: both services (`orders`, `products`), `shared` module, protos, build tooling.

## Summary

This is a Go workspace (`go.work`) with two microservices sharing a consistent
handler → service → repository architecture, gRPC for inter-service calls, and
a `shared` logging module. For a learning project it is in good shape: layering
is consistent, dependencies are injected as interfaces, error handling is
centralized, and most layers have decent test coverage. The main gaps are
around production-readiness (config, graceful shutdown, error fidelity across
the gRPC boundary) and a couple of untested/duplicated areas — expected at this
stage (see [TECH.md](TECH.md) phases 3, 4, 7 which cover Postgres, config, and
Docker and are not yet implemented).

## Strengths

- **Consistent layering** — both services follow handler → service →
  repository with constructor injection (`NewXxx`) and interfaces owned by the
  consumer (e.g. `OrderRepository` defined in [order_service.go](services/orders/internal/service/order_service.go), not in the repository package). This is idiomatic Go and keeps the domain decoupled from storage.
- **Correct concurrency handling** — the in-memory repositories
  ([order_repository.go](services/orders/internal/repository/order_repository.go), [product_repository.go](services/products/internal/repository/product_repository.go))
  use `sync.RWMutex` with locks on every map access (not just writes), with
  comments explaining why. This is a common beginner bug that's been avoided.
- **Server-generated IDs** — both `CreateOrder`/`CreateProduct` generate
  `uuid.NewV7()` server-side rather than trusting client input, which is both
  more secure (no ID injection/collision from the client) and correct
  (monotonic, sortable IDs instead of a `len()`-based counter).
- **Clean gRPC boundary** — [products_client.go](services/orders/internal/grpc/products_client.go)
  wraps the generated `ProductServiceClient` behind a small `ProductsClient`
  interface consumed by `OrderService`, keeping generated code out of the
  service layer and making it trivially mockable in tests.
- **Correct gRPC status mapping** — [grpc/errors.go](services/products/internal/grpc/errors.go)
  translates domain sentinel errors to proper `codes.NotFound` /
  `codes.InvalidArgument` / `codes.Internal`, instead of leaking `Internal` for
  everything.
- **Centralized HTTP error handling** — [error_response.go](services/products/internal/http/error_response.go)/`errors.go`
  in both services map sentinel errors and `ValidationError` to a consistent
  `{code, message}` JSON envelope and status code in one place, instead of
  scattering `http.Error` calls through handlers.
- **Structured, request-scoped logging** — `shared/logger` wraps zerolog,
  bridges to `slog`, and [request_context_middleware.go](shared/logger/request_context_middleware.go)
  attaches a request-scoped logger to context, retrieved via `log(ctx)` in
  handlers. Good separation between infra (`shared`) and per-service usage.
- **Decent test coverage** — handler, service, domain, and validation packages
  all have tests (`_test.go` alongside the code they cover), using
  table-driven style with testify. `make test-all` is green across all three
  modules.

## Findings

### High priority

1. **Repository layer has zero test coverage.**
   Neither [order_repository.go](services/orders/internal/repository/order_repository.go)
   nor [product_repository.go](services/products/internal/repository/product_repository.go)
   has a `_test.go` file. This is the one layer with real concurrency logic
   (the `RWMutex`), and it's exactly the kind of code a `go test -race` run
   should be exercising with concurrent goroutines. Currently nothing verifies
   the locking actually prevents the race it's meant to prevent.

2. **gRPC error information is discarded at the client boundary.**
   In [order_service.go](services/orders/internal/service/order_service.go),
   `validateProductExists` collapses *every* error from the products gRPC
   call — `NotFound`, `Unavailable`, `DeadlineExceeded`, `InvalidArgument` —
   into `ErrProductNotFound` (already flagged with a `TODO` in the code). If
   the products service is down or slow, orders currently reports "product not
   found" (404) instead of "upstream unavailable" (502/503/504), which is
   misleading to API consumers and hides real outages in logs/metrics.

3. **No graceful shutdown.**
   Both `cmd/main.go` entry points call `http.ListenAndServe` /
   `grpcServer.Serve` directly with no `signal.NotifyContext`, no
   `http.Server` with `Shutdown(ctx)`, and no shutdown timeout. In-flight
   requests are dropped on deploy/restart, and there's no way to drain
   connections cleanly.

4. **Bind/listen errors are silently swallowed.**
   In [products/cmd/main.go](services/products/cmd/main.go), the HTTP server
   runs in a bare `go func() { http.ListenAndServe(...) }()` — if the port is
   already in use, the goroutine exits silently and the process keeps running
   with only gRPC alive, giving no indication the HTTP API never started.

### Medium priority

5. **No externalized configuration.**
   Ports and addresses (`:8082`, `:8092`, `localhost:8092`, `:8083`) are
   hardcoded in `main.go`. This blocks running multiple instances, testing
   against different environments, or containerizing without source edits.
   (Tracked as Phase 4 in [TECH.md](TECH.md), not yet started.)

6. **Duplicated code between services instead of living in `shared`.**
   `validation/uuid.go` (`GetValidUUID`/`ErrInvalidUUID`) is duplicated
   verbatim in both `orders` and `products`. The repository mutex/CRUD
   boilerplate is likewise near-identical between
   `order_repository.go`/`product_repository.go`. These are good candidates
   to extract into `shared` once the pattern stabilizes (e.g. a generic
   `InMemoryRepository[K, V]`).

7. **Ignored return values on response encoding.**
   Handlers call `json.NewEncoder(w).Encode(...)` without checking the
   returned error (e.g. [order_handler.go](services/orders/internal/http/order_handler.go)).
   In practice this rarely fails after headers are written, but it's worth at
   least logging the error rather than discarding it silently.

8. **Package naming inconsistency.**
   `products`'s HTTP package is named `httpx` while `orders`'s equivalent
   package is literally named `http`, shadowing the standard library package
   name inside that file. Harmless today but inconsistent and a bit
   surprising to a reader jumping between services — worth aligning on
   `httpx` everywhere.

9. **`products/cmd/main.go` mixes scratch code with server bootstrap.**
   Manual map lookups, discount math, and console-only demo calls
   (`product1.ApplyDiscount(10)`, `products["999"]`) live directly in `main`
   alongside real server wiring. Fine for an intentional learning sandbox, but
   worth ring-fencing (e.g. behind a `-demo` flag or moved to an example file)
   before it's mistaken for production bootstrap code.

### Low priority

10. **Stray `coverage.out` at the repo root.**
    `make clean` only removes `coverage-shared.out`, `coverage-orders.out`,
    `coverage-products.out` — a plain `coverage.out` exists at the root and
    isn't covered by the `clean` target (it *is* gitignored via
    `coverage*.out`, so this is just local hygiene, not a repo risk).

11. **No CI pipeline.**
    There's no `.github/workflows` (or equivalent), so `make test-all` /
    `make lint` only run when a developer remembers to run them locally.

12. **No `golangci-lint` config.**
    `make lint` calls `golangci-lint run` per module, but there's no
    `.golangci.yml` pinning enabled linters/rules, so results depend on
    whatever the invoking machine has installed/configured as default.

## Alignment with the learning roadmap ([TECH.md](TECH.md))

| Phase | Topic | Status |
|---|---|---|
| 1 | Go fundamentals | Done |
| 2 | HTTP API (chi) | Done |
| 3 | PostgreSQL integration | Not started (in-memory repos only) |
| 4 | Configuration (env vars) | Not started (hardcoded ports) |
| 5 | Structured logging | Done (`shared/logger`, zerolog + slog) |
| 6 | Testing | Mostly done — repository layer untested (finding #1) |
| 7 | Docker / Compose | Not started |
| 8 | gRPC | Partially done — `GetProductByID` only, no `CreateProduct` over gRPC |
| 9 | Order service | Done |
| 10 | REST client (switchable transport) | Not started |
| 11 | Context propagation / timeouts | Partially done — `context.Context` threaded through, but no deadlines/timeouts configured on either HTTP or gRPC clients |
| 12–13 | Kafka events | Not started |
| 14 | Concurrency (goroutines/channels/worker pools) | Partially done — mutex-based safety only, no goroutines/channels/worker pools yet |

All findings above, plus a few extra items, are tracked as checkboxes in
[open-tasks.md](open-tasks.md).

## Suggested next steps (priority order)

1. Add repository-layer tests, including a `-race` concurrent test hitting
   `Save`/`Update`/`Delete`/`FindAll` from multiple goroutines.
2. Fix the gRPC error-collapsing TODO in `OrderService.validateProductExists`
   — map `codes.NotFound` → `ErrProductNotFound`, everything else →
   a new `ErrProductServiceUnavailable`/`ErrUpstream` mapped to 502/503.
3. Add graceful shutdown (`signal.NotifyContext` + `http.Server.Shutdown` +
   `grpcServer.GracefulStop`) to both `cmd/main.go` files.
4. Externalize ports/addresses via env vars (Phase 4), which also unblocks
   Docker Compose (Phase 7).
