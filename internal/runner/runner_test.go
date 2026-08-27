package runner

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitStopsOnCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if err := wait(ctx, time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait error = %v, want context.Canceled", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("cancelled wait took %s", elapsed)
	}
}
