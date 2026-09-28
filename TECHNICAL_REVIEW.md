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

# Technical Review

Reviewed: 2026-09-28. Fresh static review of orders, products, shared packages,
tests, SQL migrations, configuration, protobuf contract, containers, and tooling.
No builds, tests, databases, or containers were run; test results are not certified.

## Assessment

The structure is maintainable, but correctness and failure handling are not yet
strong enough for production ownership. Invest next in reliable behavior under
invalid input, database failures, cancellation, and concurrent updates, not
additional infrastructure or architectural layers.

Preserve the useful foundations: parameterized SQL, context-aware repositories,
consumer-owned interfaces, wrapped errors, centralized transport responses,
database-generated timestamps, embedded migrations, request-ID propagation,
non-root images, and the vet/race/formatting checks in `make check`.

## Current Defects

Ordered by remediation priority; these describe current code, not roadmap omissions.

### 1. Product deletion reports success after failure

[admin_handler.go](services/products/internal/http/admin_handler.go#L108)
discards `DeleteProduct`'s error and always sends 204. Handle the error before
writing the response. Add a handler test injecting a delete failure and expecting
500; [existing tests](services/products/internal/http/admin_handler_test.go#L475)
cover successful deletion and malformed IDs, not this failure.

### 2. Category reads have two correctness bugs

[product_category_repository_postgre.go](services/products/internal/repository/product_category_repository_postgre.go)
uses `ILIKE` for identity lookup: `%` and `_` are patterns, not literal names.
Validation can accept `%`, return it as the category, and then fail the product
foreign key with a 500. Compare exact canonical database names; keep categories
database-driven, not a Go enum.

`GetAll` also omits `rows.Err()`, turning iteration failures into successful
partial responses. Its [iteration-error test](services/products/internal/repository/product_category_repository_postgre_test.go#L99)
expects an error the implementation does not return. Check the terminal error
and explicitly choose whether repository errors return nil or partial results.

### 3. Update races become internal errors

[Orders](services/orders/internal/repository/order_repository_postgre.go) and
[products](services/products/internal/repository/product_repository_postgre.go)
translate `pgx.ErrNoRows` in `FindByID`, but not in `Update`. Both services read
before updating; deletion between those operations becomes 500 instead of
not-found. Translate missing rows consistently at repository and service
boundaries; test a successful read followed by an update miss. Separately, decide
whether concurrent edits are last-write-wins or require optimistic concurrency.
A transaction alone does not define that policy.

### 4. Invalid input and dependency failures are misclassified

[Order validation](services/orders/internal/http/dto_validation.go) checks only
that `product_id` is nonblank. A malformed UUID reaches products, becomes gRPC
`InvalidArgument`, and returns HTTP 500 through
[order_service.go](services/orders/internal/service/order_service.go) and
[HTTP error mapping](services/orders/internal/http/errors.go). Validate before
the RPC and return 400.

Not-found is preserved correctly; unavailable and deadline errors still lack
an explicit public policy. Map dependency failures deliberately, for example
503/504, preserving causes internally. Products'
[gRPC mapper](services/products/internal/grpc/errors.go) should recognize wrapped
context cancellation/deadline errors rather than defaulting to Internal.
Test each classification, not just that an error exists.

### 5. Request work is insufficiently bounded

[Products main](services/products/cmd/main.go) and
[orders main](services/orders/cmd/main.go) construct HTTP servers without timeout
settings. The [products client](services/orders/internal/grpc/products_client.go)
has no call budget; startup database Ping has no application deadline. Configure
header/idle timeouts, deliberate read/write limits, and request/dependency/startup
budgets preserving earlier caller deadlines. A server write timeout does not
replace cancellation of database/RPC work.

POST/PUT handlers decode once without a body limit or EOF check: oversized bodies
consume resources, and trailing JSON is silently accepted. Enforce body limits
and one JSON value; explicitly choose an unknown-field policy.

### 6. Shutdown can skip resources

In [products main](services/products/cmd/main.go), an HTTP shutdown error returns
before gRPC is stopped. `Shutdown` timing out does not force active HTTP
connections closed. Attempt every cleanup, force-close after the drain budget
where necessary, and aggregate errors. Preserve drain-before-pool-close ordering.

Server goroutines call `Fatal`, bypassing deferred cleanup. Report serve failures
to one lifecycle owner. The [gRPC wrapper](services/orders/internal/grpc/products_client.go)
also loses connection ownership: retain and close the `ClientConn`. Its `MustNew`
should not terminate the process from an infrastructure package. Test bind failure,
drain success, deadline expiry, and cleanup after one component fails.

### 7. Validation and storage contracts disagree

[Product validation](services/products/internal/http/dto_validation.go) accepts
names/prices beyond SQL `VARCHAR(100)` / `NUMERIC(10,2)` bounds; order quantity can
exceed PostgreSQL `INT`. Such input can become a database-driven 500. Align API
and storage ranges. [Search maxPrice](services/products/internal/http/product_handler.go#L67)
also accepts `NaN`, infinities, and negative values; require a finite valid bound.

Core invariants are mostly HTTP-only: direct service calls can persist invalid
quantity, price, name, or status. Enforce business invariants below HTTP, retain
transport parsing in handlers, and add database CHECK constraints for invariants
that must survive every writer.

Money is `float64` in Go and `double` in [protobuf](protos/product/v1/product.proto),
but fixed-scale decimal in SQL. Define currency, precision, and rounding, then
use an exact representation across boundaries. `ApplyDiscount` needs an accepted
percentage range and rounding behavior before business use.

### 8. Database configuration is fragile

[postgre.go](shared/database/postgre.go) and [migration.go](shared/database/migration.go)
interpolate credentials into URLs without escaping. URI-special characters can
change parsing. Construct pgx configuration structurally and use an escaped URL
for migrations. Migrations force `sslmode=disable`; configure consistent TLS policy.

[config.go](shared/config/config.go) lacks strict YAML field checking and post-load
validation. Validate ports, positive timeouts, required addresses, and database
settings. Add runtime secret overrides: changing Compose's `POSTGRES_PASSWORD`
does not change embedded application credentials. Test special-character
credentials and invalid config.

### 9. Lists and search load whole tables

[product_service.go](services/products/internal/service/product_service.go)
loads every product and filters in Go. Orders/products lists are unbounded and
lack stable SQL ordering. Move filtering and bounded pagination into SQL with
deterministic ordering and a documented API contract. Choose indexes from query
plans, not speculative full-text/vector search requirements.

### 10. Logging has hidden global state and inaccurate status capture

[logger.New](shared/logger/logger.go) changes global level, caller formatting,
and sometimes time formatting. Creating another logger can alter the first;
the side-effect-free constructor comment is incorrect. Set levels per logger
and make unavoidable process-wide setup explicit and one-time.

[request_context_middleware.go](shared/logger/request_context_middleware.go)
records every `WriteHeader`, even after a different status was committed, and
does not track implicit writes. It also hides optional writer capabilities.
Use chi's established response-writer wrapper; test implicit 200, repeated
headers, and required streaming/controller behavior. Define panic recovery and
completion logging at HTTP and gRPC boundaries.

Response helpers still discard `w.Write` errors after marshaling; log these
without sending a second response. Replace whole request DTO Info logs with
selected fields to control volume and future sensitive-data exposure.

## Go Design And Tests

- **Keep abstractions small.** Existing repository interfaces are useful. Avoid
   generic repositories, DI frameworks, or sharing trivial duplicates solely to
   remove repetition. Concrete handler dependencies are not inherently wrong.
- **Finish the transport boundary.** `ProductsClient` returns generated protobuf
   types and `OrderService` imports gRPC status codes. Define the small application
   lookup/result/error contract needed; translate in the adapter. Repository
   sentinels also tie services to storage packages; clarify error ownership without
   introducing a generic error framework.
- **Handle ID generation errors.** Both create services ignore `uuid.NewV7()`
   errors. Return the error before persistence; do not save a zero ID.
- **Separate test responsibilities.** Service and HTTP tests repeatedly wire
   repositories to pgxmock and duplicate exact SQL. Keep selected composition tests,
   but test service decisions through repository stubs and HTTP contracts through
   narrow dependencies where useful. SQL expectations belong primarily in repository
   tests. Removing in-memory storage does not require SQL mocks in every layer.
- **Repair misleading assertions.** Use `require` before dereferencing responses
   or inspecting typed errors; prefer `ErrorIs`/`ErrorAs` over asserting every wrapped
   error string. `assert.NotNil` on a value UUID does not check it is nonzero. Audit
   `RowError(1, ...)` cases without a second row: explicitly distinguish scan,
   iteration, and no-row errors. Remove the pgx v4 import in
   [category tests](services/products/internal/service/product_category_service_test.go)
   and tidy the resulting unused dependency.
- **Add real PostgreSQL coverage.** pgxmock cannot validate SQL execution, codecs,
   rounding, constraints, or migrations. Use isolated disposable databases and
   actual migrations for CRUD, wildcard cases, constraints, and cancellation.
   Test fresh migrations, no-change reruns, and upgrades; never rewrite applied
   migrations. The category seed down-migration can fail when products reference
   its rows: document/test the supported rollback policy.
- **Exercise real boundaries selectively.** Add an orders-to-products gRPC suite
   for status codes, deadlines, and request IDs. Shutdown tests using cancellation,
   ephemeral listeners, and a focused subprocess signal test are practical; the
   previous blanket dismissal was incorrect.

## Gaps After Correctness

- **Authorization before exposure.** Product admin writes and order operations
   have no authentication/authorization. Define caller identity, admin permissions,
   and order ownership. Plaintext gRPC, default credentials, and exposed DB ports
   require a production TLS/secrets/network policy. Local defaults are not evidence
   of leaked production secrets.
- **Readiness and operations.** `/health` is liveness only. Add bounded readiness
   including shutdown state and necessary dependencies, without coupling liveness
   to transient DB failures. Add service health checks, aligned termination grace
   periods, request/error/latency and pool metrics, alerts, and basic runbooks.
   Request IDs help correlation but do not replace tracing.
- **Reproducible builds and CI.** [Dockerfiles](services/products/Dockerfile) copy
   ignored prebuilt binaries; [Makefile](Makefile) hardcodes Linux ARM64. Clean
   checkout image builds need a separate build step, and AMD64 images can receive
   incompatible binaries. Use target-aware multi-stage builds or an explicit
   artifact pipeline. CI should run `make check`, isolated DB tests, image builds,
   dependency scanning, and pinned protobuf regeneration checks. Two generated
   copies are workable if drift is checked.
- **Migration ownership.** Startup migration is convenient locally. Define
   production orchestration, schema/runtime privileges, rollout compatibility,
   and dirty-migration recovery. Library locking is not a deployment procedure.
- **Commerce semantics.** Orders retain a product reference, quantity, and
   caller-updatable status. Define legal transitions, price/currency snapshots,
   deletion semantics, idempotent creation, and inventory consistency before
   payments/retries. Product-existence RPCs are not cross-service transactions.
- **API/setup documentation.** Document REST contracts and error codes.
   [README.md](README.md) still describes in-memory products, Postgres-only Compose,
   automatic test-profile selection, incorrect default profile/log level, and
   nonexistent `make lint`. Correct the runnable setup instructions and the
   early-return shutdown example in [shared/README.md](shared/README.md).

Kafka, Redis, Kubernetes manifests, another transport, and additional RPCs are
not prerequisites for these fixes. See [open-tasks.md](open-tasks.md).
