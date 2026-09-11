package shutdown

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopGracefullyOrForcefully_WhenStopFinishesBeforeDeadline_DoesNotForceStop(t *testing.T) {
	forceStopCalled := false
	stop := func() {}
	forceStop := func() { forceStopCalled = true }

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	StopGracefullyOrForcefully(ctx, stop, forceStop)

	assert.False(t, forceStopCalled)
}

func TestStopGracefullyOrForcefully_WhenStopDoesNotFinishBeforeDeadline_CallsForceStop(t *testing.T) {
	release := make(chan struct{})
	stop := func() { <-release }
	forceStop := func() { close(release) }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		StopGracefullyOrForcefully(ctx, stop, forceStop)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("StopGracefullyOrForcefully did not return after forceStop unblocked stop")
	}
}

func TestStopGracefullyOrForcefully_WhenAlreadyDone_StillWaitsForStopToReturn(t *testing.T) {
	stopReturned := false
	stop := func() { stopReturned = true }
	forceStop := func() {}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	StopGracefullyOrForcefully(ctx, stop, forceStop)

	assert.True(t, stopReturned)
}

func TestWaitForTerminationAndShutdown_WhenParentIsCanceled_RunsShutdownWithDeadline(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	shutdownStarted := make(chan struct{})

	go func() {
		cancelParent()
	}()

	err := WaitForTerminationAndShutdown(parent, time.Second, func(ctx context.Context) error {
		_, hasDeadline := ctx.Deadline()
		assert.True(t, hasDeadline)
		close(shutdownStarted)
		return nil
	})

	assert.NoError(t, err)
	select {
	case <-shutdownStarted:
	default:
		t.Fatal("shutdown callback was not called")
	}
}
