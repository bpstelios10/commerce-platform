# Open Tasks

Prioritized from the fresh 2026-09-28 [technical review](TECHNICAL_REVIEW.md).
Only unfinished work is listed. Acceptance checks are proposed, not executed.

## 1. Fix Current Defects

- [x] Handle product delete errors in HTTP; test DB failure returns 500 rather
      than 204, while successful/idempotent deletion remains 204.
- [x] Use literal category lookup against canonical database names; test `%`,
      `_`, unknown categories, and normalized valid names against PostgreSQL.
- [x] Check category `rows.Err()`; verify iteration failures cannot return success.
      Make partial-result behavior explicit and consistent where appropriate.
- [x] Translate UPDATE no-rows into repository and service not-found errors in
      both services; test deletion between the preliminary read and write.
- [x] Validate order product UUIDs before RPC.
- [x] Handle UUIDv7 generation errors in both create services before persistence.
- [ ] Align input bounds with SQL name/price/quantity limits; reject nonfinite or
      invalid search price bounds. Enforce core invariants below HTTP and add
      database CHECK constraints through new migrations.
- [x] Limit HTTP request bodies, require one JSON value, and choose/test the
      unknown-field policy. Cover oversized and trailing-content requests.
- [ ] Configure HTTP timeouts and request, RPC, database, and startup budgets.
      Verify cancellation reaches dependencies and earlier deadlines are preserved.
- [ ] Attempt all shutdown cleanup even after an error; force-close after the
      drain budget. Centralize serve failures instead of `Fatal` in goroutines;
      retain and close the products gRPC connection.
- [ ] Use escaped/structured DB configuration, consistent TLS, runtime secret
      overrides, and startup validation. Test URI-special credentials, invalid
      ports/timeouts, and unknown YAML fields.

## 2. Improve Go And API Quality

- [ ] Define exact money/currency/rounding contracts across Go, JSON, protobuf,
      and PostgreSQL. Validate discount ranges and test rounding boundaries.
- [ ] Move product search filtering into SQL; add bounded pagination and stable
      ordering to orders/products lists. Test filters, limits, and page continuity.
- [ ] Define an application-owned product lookup contract; keep generated types
      and gRPC status translation in the adapter. Clarify repository error
      ownership without introducing a generic framework.
- [ ] Make logger levels instance-local and global formatting setup explicit.
      Test that creating one logger does not change another's level.
- [ ] Use a proven HTTP response-writer wrapper; test implicit/repeated statuses
      and required capabilities. Define/test HTTP and gRPC panic recovery.
- [ ] Log response-write failures without a second response; replace whole DTO
      Info logs with selected fields. Preserve context without duplicate noise.
- [ ] Decide concurrent-update semantics; use optimistic concurrency where lost
      updates violate the intended contract.

## 3. Strengthen Verification

- [x] Focus pgxmock on repository behavior; use repository stubs for service
      decisions and isolate HTTP contracts where helpful. Keep selected composition
      tests instead of duplicating SQL throughout every layer.
- [ ] Use fatal test prerequisites, meaningful UUID assertions, and error identity
      checks. Audit row-error fixtures so each exercises its named failure; remove
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

- [ ] Make clean-checkout image builds self-contained and target-aware for
      ARM64/AMD64; distinguish native builds from container artifacts.
- [ ] Correct setup/API docs: PostgreSQL repositories, all Compose services,
      explicit profiles, log-level precedence, `make check`, credential overrides,
      and the shutdown example. Document REST contracts and errors.
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

## Defer

Kafka, Redis, Kubernetes deployment, alternate transports, extra RPCs, generic
repositories, and advanced search infrastructure until a concrete use case needs
them. Keep learning-roadmap experiments separate from this queue.
