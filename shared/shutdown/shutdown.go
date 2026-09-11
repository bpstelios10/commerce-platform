package shutdown

import (
	"context"
	"os/signal"
	"syscall"
	"time"
)

func ContextWithTerminationSignals(parent context.Context) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
}

func WaitForTerminationAndShutdown(
	parent context.Context, timeout time.Duration, shutdown func(context.Context) error) error {

	ctx, stop := ContextWithTerminationSignals(parent)
	defer stop()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return shutdown(shutdownCtx)
}

func StopGracefullyOrForcefully(ctx context.Context, stop func(), forceStop func()) {
	done := make(chan struct{})

	go func() {
		stop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		forceStop()
		<-done
	}
}
