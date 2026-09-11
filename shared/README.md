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
