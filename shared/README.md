# Shared Packages

## Graceful Shutdown

The `shutdown` package contains reusable shutdown mechanics for services.

```go
if err := shutdown.WaitForTerminationAndShutdown(
    context.Background(),
    10*time.Second,
    func(shutdownCtx context.Context) error {
        if err := httpServer.Shutdown(shutdownCtx); err != nil {
            return err
        }

        shutdown.StopGracefullyOrForcefully(
            shutdownCtx,
            grpcServer.GracefulStop,
            grpcServer.Stop,
        )
        return nil
    },
); err != nil {
    logger.Error().Err(err).Msg("service shutdown failed")
}
```

### What it does

- `WaitForTerminationAndShutdown` waits for `SIGINT` or `SIGTERM`, creates the bounded shutdown context, and
    runs the service-specific shutdown callback.
- `http.Server.Shutdown` stops new HTTP work and waits for active requests to finish.
- `StopGracefullyOrForcefully` waits for graceful shutdown, then calls the force-stop function when the deadline expires.
- `WaitForTerminationAndShutdown` releases the signal and timeout resources when the service exits.

The service `main.go` files keep the shutdown order visible. This matters because
HTTP, gRPC, workers, message consumers, and database connections may need to stop
in different orders.

## Configuration

The `config` package loads YAML config with profile overlays.

```go
//go:embed base.yaml local.yaml test.yaml
var configFiles embed.FS

var cfg MyConfig
err := config.Load(configFiles, os.Getenv("ACTIVE_PROFILE"), &cfg)
```

- `base.yaml` is always loaded first.
- If a profile is set (e.g. `local`, `test`), `<profile>.yaml` is loaded on top of it,
  overriding only the fields it sets.
- `Load` takes an `fs.FS`, not a concrete `embed.FS`, so tests can pass an in-memory
  `fstest.MapFS` instead of real files on disk.
- Each service embeds and owns its own config files; this package only knows how to
  load and overlay them.

## Database (PostgreSQL)

The `database` package wraps `pgx`/`golang-migrate` for the two services that use Postgres.

```go
pool, err := database.NewPostgreClient(ctx, cfg.Database)
// ...
err = database.RunMigrations(cfg.Database, migrationFiles)
```

- `NewPostgreClient` builds a connection pool from a `DatabaseConfig`; it doesn't
  connect eagerly, so callers should `Ping` it themselves to fail fast on startup.
- `RunMigrations` runs a service's own embedded `.sql` migrations
  (`services/<name>/migrations`) against its own database, using `golang-migrate`.
- Each service owns its schema and its own migration files; this package only knows
  how to connect and how to run whatever migrations it's given.

