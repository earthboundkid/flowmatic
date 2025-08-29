package flowmatic_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/earthboundkid/flowmatic/v2"
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
	synctest.Test(t, func(t *testing.T) {
		var (
			a = errors.New("a")
			b = errors.New("b")
		)
		err := flowmatic.Race(context.Background(),
			func(ctx context.Context) error {
				if !sleepFor(ctx, 1*time.Second) {
					return ctx.Err()
				}
				return a
			},
			func(ctx context.Context) error {
				if !sleepFor(ctx, 3*time.Second) {
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
	})
}
