# Technical Review

Reviewed: 2026-10-31. Static review of both services, `shared`, migrations,
config, proto and build files. Nothing was built, tested or run. Supersedes the
2026-09-28 review.

## Assessment

Clear progress since September: optimistic concurrency, structured pgx config,
request/RPC/DB timeouts, body limits and single-value JSON decoding, and a much
better test split (fakes with call verification, composition tests, pgxmock only
in repositories). Layering is sound; keep it.

What now separates this from production quality is data integrity at the DB
layer, process lifecycle, unvalidated config and some non-idiomatic naming.
Item 1 is a blocker for any existing database.

## Current Defects

Ordered by priority. Several were marked done in [open-tasks.md](open-tasks.md)
but are not done in code; the task list is corrected.

### 1. An applied migration was edited (blocker)

`version` was added inside `001_create_orders.up.sql` and
`001_create_products.up.sql`. golang-migrate already recorded version 1 on any
existing database (including the Compose volume), so the column is never added
and every query that selects `version` fails. Revert both `001` files and add
new migrations (`ALTER TABLE ... ADD COLUMN version INT NOT NULL DEFAULT 1`).
Never edit an applied migration; a "fresh DB + upgrade from previous version"
test catches this.

### 2. Integrity rules exist only in HTTP handlers

- No CHECK constraints exist in any migration.
- Services accept `quantity <= 0`, empty names, non-positive prices and any
  status-to-status transition.
- Product name/category length is validated on the trimmed value, but the
  untrimmed value is persisted, so padded input can exceed `VARCHAR(100)` and
  return 500.

Trim before validating and storing. Enforce invariants in the domain/service
layer (constructors or `Validate()` on the domain type). Add a new migration with
CHECKs: `quantity > 0`, `price > 0`, `status IN (...)`, `version > 0`.

### 3. Category `List` still ignores iteration errors

[product_category_repository_postgre.go](services/products/internal/repository/product_category_repository_postgre.go)
has no `rows.Err()` check and returns a partial slice on scan errors. A
truncated result is reported as success.

### 4. Process lifecycle

- Serve goroutines in both `main.go` files call `logger.Fatal`, which skips the
  deferred `db.Close()`.
- In products, an HTTP `Shutdown` error returns before gRPC is stopped.
- Orders never closes the products `ClientConn`, and
  `MustNewProductsGrpcClient` now panics instead of returning an error to `main`.

Let `main` own the lifecycle: start servers under `errgroup` (or a shared error
channel), return errors from constructors, close the gRPC connection on
shutdown, attempt every cleanup step and `errors.Join` the results.

### 5. Config is unvalidated, and some values bypass it

- No strict decoding (`yaml.Decoder.KnownFields(true)`) and no `Validate()`.
  A missing timeout is zero, and `context.WithTimeout(ctx, 0)` fails every query
  immediately. The `uint16`/`int32` casts of port and pool sizes overflow
  silently.
- The products gRPC call timeout is hardcoded (`500*time.Millisecond`) in
  [orders main](services/orders/cmd/main.go).
- TLS is inconsistent: the pool uses pgx's default `sslmode` (`prefer`), while
  migrations force `disable`. Make `sslmode` a config field used by both.
- The migration `connect_timeout` truncates sub-second values to `0`, which
  Postgres treats as no timeout.
- `GracefulShutdown.Timeout` is int seconds; every other duration is
  `time.Duration`.

### 6. Money is still floating point

Prices are `float64` in Go, `double` in the proto and `NUMERIC(10,2)` in SQL.
`ApplyDiscount` multiplies floats without rounding, so results drift from what
Postgres stores. Use integer minor units (`int64` cents) across Go, proto and
JSON, or a decimal type, and define the rounding rule.

### 7. Timeout and cancellation mapping is half done

- Products' [gRPC `HandleError`](services/products/internal/grpc/errors.go)
  maps a context deadline/cancel to `Internal`. Orders then sees an internal
  error, not a timeout. Map them to `DeadlineExceeded`/`Canceled`.
- HTTP returns 503 for every `DeadlineExceeded`. 504 is the precise code for an
  upstream timeout. An oversized body returns 400; 413 is the standard code.
- `Delete` uses the `Read` timeout in both repositories.
- An UPDATE that matches no rows is always reported as a version conflict, so an
  entity deleted between the read and the write returns 409, not 404. Fine if
  intended; document it.

### 8. Lists and search load whole tables

Unchanged: `SearchProducts` filters the full table in Go, and lists have no
pagination or ordering. Move filtering, `ORDER BY` and a `LIMIT`/cursor into SQL.

### 9. Logging global state and status capture

Unchanged: [`logger.New`](shared/logger/logger.go) mutates the global level and
format functions, and its comment still claims it is side-effect-free. The
[middleware](shared/logger/request_context_middleware.go) records every
`WriteHeader` call, even after the status is committed. There is no panic
recovery at the HTTP or gRPC boundary; add `middleware.Recoverer` and a gRPC
recovery interceptor.

## Idiomatic Go

- **Drop the `Impl` suffix.** `OrderServiceImpl`, `ProductServiceImpl` and
  `AdminServiceImpl` are a Java pattern. Name the concrete type
  `service.OrderService` and keep the consumer's interface in the handler
  package, unexported. Constructors are inconsistent too: `NewOrderService`
  versus `NewProductServiceImpl`.
- **Depend on behavior, not concrete services.** `AdminServiceImpl` holds
  `*ProductServiceImpl` only to call `GetProductByID`. Depend on a one-method
  interface or on the repository.
- **Error naming.** Sentinels are `ErrX`; error types are `XError`
  (`OrderNotFoundError`, `BodyTooLargeError`, not `ErrOrderNotFound` or
  `TooLongBodyErr`). Error strings are lowercase without punctuation
  (`ApplyDiscount`, `TooLongBodyErr`). Add `staticcheck` to `make check`; it
  flags these.
- **Body decoding helper.** `isBodyValidAndDecoded` hardcodes `ErrInvalidOrder`
  and `1<<10`, and decodes into `&req` where `req` is already `*T`. Pass `req`,
  take the limit and error as parameters, and name it `decodeJSONBody`.
- **Error context.** Prefixes mix styles: `get_order:` (handler),
  `get order by id:` (service), `find order by id:` (repo `GetByID`). Use one
  style and match the function name.
- **Proto.** Package `product` is unversioned; use `product.v1` and run
  `buf lint`/`buf breaking` in CI.
- **Transport boundary.** Unchanged: `OrderService` still imports generated
  protobuf types and gRPC status codes. Translate both in the gRPC adapter.

## Tests

The current split is right: fakes with expected-call verification, one
composition file per handler, and pgxmock only in repositories. The main gap is
real PostgreSQL: items 1, 2 and 3 would each have failed a migration-backed
integration test. Also missing: a gRPC client deadline test and lifecycle tests
(bind failure, drain, cleanup after a failure).

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
   deletion semantics, and inventory consistency before payments. With timeouts
   in place, a client can no longer tell whether a timed-out create committed:
   add an `Idempotency-Key` on `POST /orders` before clients retry.
   Product-existence RPCs are not cross-service transactions.
- **API/setup documentation.** Document REST contracts and error codes.
   [README.md](README.md) still describes in-memory products, Postgres-only Compose,
   automatic test-profile selection, incorrect default profile/log level, and
   nonexistent `make lint`. Correct the runnable setup instructions and the
   early-return shutdown example in [shared/README.md](shared/README.md).

Kafka, Redis, Kubernetes manifests, another transport, and additional RPCs are
not prerequisites for these fixes. See [open-tasks.md](open-tasks.md).
