package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleRace() {
	sleepFor := func(d time.Duration) func(context.Context) error {
		return func(ctx context.Context) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-timer.C:
				fmt.Println("timer:", d)
				return nil
			case <-ctx.Done():
				fmt.Println("cancelled")
				return ctx.Err()
			}
		}
	}
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.Race(ctx,
		sleepFor(1*time.Minute),
		sleepFor(1*time.Second),
		sleepFor(1*time.Millisecond),
	)
	fmt.Println(err, time.Since(start).Round(time.Second))
	// Output:
	// timer: 1ms
	// cancelled
	// cancelled
	// <nil> 0s
}
