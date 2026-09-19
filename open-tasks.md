# Open Tasks

Tracks the follow-ups from [TECHNICAL_REVIEW.md](TECHNICAL_REVIEW.md), plus a
few extra items worth doing. Check items off as they're done; update
[TECH.md](TECH.md)'s status table when a phase-related item lands.

## High priority

- [ ] Stop collapsing all gRPC errors into `ErrProductNotFound` in
      `OrderService.validateProductExists` ([order_service.go](services/orders/internal/service/order_service.go)).
      Map `codes.NotFound` → `ErrProductNotFound`; map everything else (`Unavailable`,
      `DeadlineExceeded`, etc.) to a new `ErrProductServiceUnavailable` → HTTP 502/503.
- [x] Add graceful shutdown to both `cmd/main.go` entry points: `signal.NotifyContext`
      (SIGINT/SIGTERM) + `http.Server.Shutdown(ctx)` + `grpcServer.GracefulStop()`, with a
      bounded shutdown timeout (`shutdownTimeout = 10s`; gRPC falls back to a forceful
      `Stop()` if `GracefulStop()` doesn't finish in time, via the tested
      [shared/shutdown.StopGracefullyOrForcefully](shared/shutdown/shutdown.go)). Manually
      verified with `kill -TERM` against both running binaries — clean exit, no dropped/hung
      connections. (Full signal-handling flow in `main()` itself is intentionally not unit
      tested — extracting real signals/ports into a test is unconventional/fragile; the
      testable timeout-fallback logic is covered instead.)
- [x] Fix silent HTTP bind failures in [products/cmd/main.go](services/products/cmd/main.go) —
      the HTTP server ran in a bare `go func(){ http.ListenAndServe(...) }()`; a bind error
      (e.g. port in use) was dropped. Fixed as part of the logging unification below: both the
      HTTP and gRPC listeners now log via `logger.Fatal().Err(...)`, which logs and exits on error.

## Medium priority

- [x] Externalize configuration: ports (`:8082`, `:8092`, `:8083`) and the orders→products
      gRPC address (`localhost:8092`) are hardcoded in `main.go`. Load from env vars with
      sane local defaults (unblocks Docker Compose too).
- [ ] Extract duplicated `validation/uuid.go` (`GetValidUUID`/`ErrInvalidUUID`) — currently
      copy-pasted in both `orders` and `products` — into `shared`.
- [ ] Consider extracting the repository mutex/CRUD boilerplate (near-identical between the
      two in-memory repos) into a generic `shared` helper, e.g. `InMemoryRepository[K, V]`.
- [x] Check/log the error returned by `json.NewEncoder(w).Encode(...)` in handlers instead of
      discarding it (e.g. [order_handler.go](services/orders/internal/http/order_handler.go)).
- [x] Align package naming: rename orders' `http` package to `httpx` (matches products, and
      stops shadowing the stdlib `net/http` import name inside the package).
- [ ] Ring-fence the scratch/demo code in [products/cmd/main.go](services/products/cmd/main.go)
      (manual map lookups, `ApplyDiscount` demo, etc.) — e.g. move behind a `-demo` flag or
      into a separate example file — so it doesn't get mistaken for real bootstrap logic.
- [ ] Add a `.golangci.yml` at the repo root so `make lint` is reproducible across machines
      instead of depending on whatever's installed locally.
- [ ] Add a CI pipeline (GitHub Actions) running `make test-all` and `make lint` on every PR.

## Low priority

- [ ] Add `coverage.out` (root-level, no prefix) to the `make clean` target — currently only
      `coverage-{shared,orders,products}.out` are removed.
- [ ] Add request timeouts: wrap HTTP servers with `http.TimeoutHandler` / set
      `ReadHeaderTimeout` etc., and set a dial/call timeout on the orders→products gRPC client
      (`context.WithTimeout` at the call site) — ties into TECH.md Phase 11 (context propagation).

## Additional suggestions (beyond the original review)

- [ ] Expose `CreateProduct` over gRPC too (currently HTTP-only), matching TECH.md Phase 8's
      goal of `GetProduct()` + `CreateProduct()` both over gRPC.
- [ ] Add a root `docker-compose.yml` once config is externalized (Phase 7) — even without
      Postgres yet, this is useful for running both services + health checks with one command.
- [ ] Add a basic OpenAPI/Swagger spec (or at least a `docs/api.md`) for the two REST APIs —
      there's currently no request/response contract documented outside of the DTO structs.
- [ ] Decide on `orders` repo's `Save` vs `Update` — they're currently identical upserts
      (`repo.orders[o.ID] = o`); either differentiate them (e.g. `Save` rejects existing IDs)
      or collapse to one method to avoid the false impression they behave differently.
- [ ] Add an integration-style test that boots both services (or fakes the gRPC boundary) to
      exercise the orders→products call path end-to-end, not just via mocked `ProductsClient`.

## Logging (Google/K8s/Uber-standard pass)

Tracked as its own section since it's a multi-step effort, done one item at a time.

- [x] **1. Kill the slog/zerolog split.** Removed all `log/slog` usage from repository,
      service, and `main.go` (both services) — everything now logs through the shared
      zerolog logger, threaded via `context.Context` (repository/service methods now take
      `ctx` as their first parameter). `shared/logger.SetAsDefault` (called once from each
      `main.go`) sets `zerolog.DefaultContextLogger` so code paths without a request-scoped
      logger in context (e.g. gRPC, until item 2 landed) fall back to the base service logger
      instead of a disabled one; kept out of `New` itself so `New` stays side-effect-free and
      safe to call repeatedly in tests.
- [x] **2. Propagate the request ID across the gRPC boundary.** Orders forwards `request_id`
      as outgoing gRPC metadata (`shared/logger.RequestIDMetadataKey`); products reads it in a
      new `LoggingUnaryInterceptor` ([interceptor.go](services/products/internal/grpc/interceptor.go))
      and injects it into its request-scoped logger, generating one if absent (mirrors the HTTP
      middleware behavior). Registered via `grpc.UnaryInterceptor(...)` in products' `main.go`.
- [x] **3. Add one canonical access-log line per request/RPC.** HTTP middleware
      ([request_context_middleware.go](shared/logger/request_context_middleware.go)) now wraps
      the `http.ResponseWriter` to capture the status code and logs one "request completed"
      line per request (method, path, status, duration). Products' `LoggingUnaryInterceptor`
      ([interceptor.go](services/products/internal/grpc/interceptor.go)) does the same for gRPC,
      logging one "rpc completed" line (method, status code, duration) per call — both separate
      from whatever business-event logs individual handlers add.
- [x] **4. Standardize log field names** across both services (snake_case: `order_id`,
      `product_id`, `request_id`, `category`) — full sweep of every `.Str`/`.Int`/`.Interface`/
      `.Bool`/`.Strs` log call across both services found field names were already snake_case
      from items 1–3, except one leftover: `Interface("maxPrice", ...)` in
      [product_handler.go](services/products/internal/http/product_handler.go)'s
      `SearchProducts`, fixed to `max_price` (the `maxPrice` HTTP query parameter itself is
      the API contract and was left unchanged).
- [x] **5. Make log level configurable via env var.** Added `shared/logger.LevelFromEnv`
      (parses a level name via `zerolog.ParseLevel`, falls back to a default if unset/invalid);
      both `cmd/main.go` files now use `loggerx.LevelFromEnv("LOG_LEVEL", zerolog.InfoLevel)`
      instead of a hardcoded level. Set `LOG_LEVEL=debug` (or `warn`/`error`) to change
      verbosity per environment without a rebuild.
- [ ] *(Deferred, own future phase)* Full OpenTelemetry trace/span propagation instead of the
      hand-rolled `request_id` — bigger lift (SDK, exporters), tracked separately in
      [TECH.md](TECH.md) rather than bundled into this pass.
