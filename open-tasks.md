# Open Tasks

Prioritized from the 2026-10-31 [technical review](TECHNICAL_REVIEW.md).
Acceptance checks are proposed, not executed.

## 0. Blocker

- [ ] Revert the `version` column out of both `001_*.up.sql` files and add it
      through new migrations. Verify: an existing DB at version 1 upgrades cleanly.

## 1. Fix Current Defects

- [x] Handle product delete errors in HTTP; test DB failure returns 500 rather
      than 204, while successful/idempotent deletion remains 204.
- [x] Use literal category lookup against canonical database names; test `%`,
      `_`, unknown categories, and normalized valid names against PostgreSQL.
- [ ] Check category `rows.Err()` and return nil on scan/iteration errors
      (still missing in `PostgreProductCategoryRepository.List`).
- [x] Translate UPDATE no-rows into repository and service not-found errors in
      both services; test deletion between the preliminary read and write.
- [x] Validate order product UUIDs before RPC.
- [x] Handle UUIDv7 generation errors in both create services before persistence.
- [x] Align HTTP input bounds with SQL name/price/quantity limits; reject
      nonfinite or invalid search price bounds.
- [ ] Trim names/categories before validating and storing (padded input can
      exceed `VARCHAR(100)`).
- [ ] Enforce invariants below HTTP (domain `Validate()` or service checks) and
      add CHECK constraints in a new migration: quantity, price, status, version.
- [x] Limit HTTP request bodies, require one JSON value, and choose/test the
      unknown-field policy. Cover oversized and trailing-content requests.
- [x] Configure HTTP timeouts and request, RPC, database
      Verify cancellation reaches dependencies and earlier deadlines are preserved.
- [x] Bound startup database connection/Ping separately, Startup has no incoming request 
      context. Database queries during requests can initially use the overall request deadline; separate query budgets are optional.
- [ ] Map context deadline/cancel in products' gRPC `HandleError` (currently
      `Internal`). Use 504 for upstream timeouts and 413 for oversized bodies.
      Use the `Write` timeout for `Delete`.
- [ ] Add an `Idempotency-Key` to `POST /orders`: a timed-out create may have
      committed, so retries can duplicate.
- [ ] Lifecycle: run servers under `errgroup` instead of `Fatal` in goroutines;
      attempt all shutdown steps and `errors.Join` them; return errors instead of
      panicking in `MustNewProductsGrpcClient`; close the products `ClientConn`.
- [x] Use structured pgx config and an escaped migration URL.
- [ ] Config: strict YAML (`KnownFields`), `Validate()` (ports, positive
      timeouts/pool sizes, required fields), `sslmode` used by pool and
      migrations, products gRPC timeout in config, shutdown timeout as
      `time.Duration`. Test invalid and unknown fields.
- [ ] Add runtime secret overrides (ENV).

## 2. Improve Go And API Quality

- [x] Use idiomatic GO repo methods (GetByID, List, Create, Update, Delete, Exists).
- [x] Use idiomatic GO capitalization everywhere (ID instead of Id).
- [ ] Use integer minor units (or a decimal type) for money across Go, proto
      and JSON; define rounding for `ApplyDiscount`.
- [ ] Drop the `Impl` suffix (`service.OrderService`), keep consumer interfaces
      unexported, and make `AdminService` depend on an interface instead of
      `*ProductServiceImpl`.
- [ ] Naming: `XError` for error types, lowercase error strings, one
      error-prefix style. Add `staticcheck` to `make check`.
- [ ] Generalize `isBodyValidAndDecoded` into `decodeJSONBody(w, r, dst, limit)`.
- [ ] Version the proto package (`product.v1`); add `buf lint`/`buf breaking`.
- [ ] Move product search filtering into SQL; add bounded pagination and stable
      ordering to orders/products lists. Test filters, limits, and page continuity.
- [ ] Define an application-owned product lookup contract; keep generated types
      and gRPC status translation in the adapter. Clarify repository error
      ownership without introducing a generic framework.
- [ ] Make logger levels instance-local and global formatting setup explicit.
      Test that creating one logger does not change another's level.
<!-- i ll leave this one for later, if ever needed -->
- [ ] Use a proven HTTP response-writer wrapper for status capture; add panic
      recovery (`middleware.Recoverer`, gRPC recovery interceptor).
- [x] Log response-write failures without a second response; 
- [x] replace whole DTO Info logs with selected fields. Preserve context without duplicate noise.
- [x] Decide concurrent-update semantics; use optimistic concurrency where lost
      updates violate the intended contract.

## 3. Strengthen Verification

- [x] Focus pgxmock on repository behavior; use repository stubs for service
      decisions and isolate HTTP contracts where helpful. Keep selected composition
      tests instead of duplicating SQL throughout every layer.
- [x] Use fatal test prerequisites, meaningful UUID assertions, and error identity
      checks. 
- [x] Audit row-error fixtures so each exercises its named failure; remove
      the stray pgx v4 import and tidy unused dependencies.
- [ ] Add isolated PostgreSQL tests for migrations, CRUD, codecs, constraints,
      category lookup, money round trips, and cancellation.
- [ ] Test fresh migrations, no-change reruns, and supported upgrades; document
      rollback limits with populated category references and dirty-migration
      recovery. Keep applied migrations immutable.
- [ ] Add orders-to-products integration tests for lookup, status mapping,
      deadlines, and request-ID propagation.
- [ ] Test lifecycle orchestration: bind failure, active-request drain, timeout
      fallback, cleanup after one failure, and a focused signal smoke test.
- [ ] Add CI running `make check` across all modules, isolated DB tests,
      dependency scanning, image builds, and pinned protobuf drift checks.

## 4. Close Deployment And Product Gaps

- [x] Build target-specific Linux binaries locally before Compose packages them;
      support ARM64 and AMD64 with `make build GOARCH=<arch>`.
- [x] Correct setup docs: PostgreSQL repositories, all Compose services,
      explicit profiles, log-level precedence, `make check`, credential overrides,
      and the shutdown example.
- [ ] Document REST contracts and errors.
- [ ] Add bounded readiness distinct from liveness, including shutdown state;
      configure service health checks and termination grace periods.
- [ ] Before external exposure, implement identity, admin permissions, order
      ownership, and an explicit TLS/secrets/network policy.
- [ ] Define production migration ownership and least-privilege runtime access;
      verify schema changes support the chosen rollout strategy.
- [ ] Add request/error/latency and DB-pool metrics, actionable alerts, and basic
      runbooks. Introduce tracing when cross-service diagnosis needs it.
- [ ] Define order transitions, price/currency snapshots, product deletion
      behavior, idempotent creation, and inventory consistency before payments.

## My Tasks
- [ ] Consider later moving validations to a GO popular library, like https://github.com/go-playground/validator?utm_source=chatgpt.com

## Defer

Kafka, Redis, Kubernetes deployment, alternate transports, extra RPCs, generic
repositories, and advanced search infrastructure until a concrete use case needs
them. Keep learning-roadmap experiments separate from this queue.
