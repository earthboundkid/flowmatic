package flowmatic_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/earthboundkid/flowmatic"
)

func sleepFor(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

func TestRace_join_errs(t *testing.T) {
	var (
		a = errors.New("a")
		b = errors.New("b")
	)

	err := flowmatic.Race(context.Background(),
		func(ctx context.Context) error {
			if !sleepFor(ctx, 10*time.Millisecond) {
				return ctx.Err()
			}
			return a
		},
		func(ctx context.Context) error {
			if !sleepFor(ctx, 30*time.Millisecond) {
				return ctx.Err()
			}
			return b
		},
	)
	if !errors.Is(err, a) || !errors.Is(err, b) {
		t.Fatal(err)
	}
	if errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
