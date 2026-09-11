package shutdown

import "context"

func GracefulStopWithTimeout(ctx context.Context, stop func(), forceStop func()) {
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
