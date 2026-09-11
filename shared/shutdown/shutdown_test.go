package shutdown

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGracefulStopWithTimeout_WhenStopFinishesBeforeDeadline_DoesNotForceStop(t *testing.T) {
	forceStopCalled := false
	stop := func() {}
	forceStop := func() { forceStopCalled = true }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	GracefulStopWithTimeout(ctx, stop, forceStop)

	assert.False(t, forceStopCalled)
}

func TestGracefulStopWithTimeout_WhenStopDoesNotFinishBeforeDeadline_CallsForceStop(t *testing.T) {
	release := make(chan struct{})
	stop := func() { <-release } // blocks until forceStop unblocks it
	forceStop := func() { close(release) }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		GracefulStopWithTimeout(ctx, stop, forceStop)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("GracefulStopWithTimeout did not return after forceStop unblocked stop")
	}
}

func TestGracefulStopWithTimeout_WhenAlreadyDone_StillWaitsForStopToReturn(t *testing.T) {
	stopReturned := false
	stop := func() { stopReturned = true }
	forceStop := func() {}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already done before GracefulStopWithTimeout is even called

	GracefulStopWithTimeout(ctx, stop, forceStop)

	assert.True(t, stopReturned)
}
