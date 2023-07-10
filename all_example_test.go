package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleDoContext() {
	sleepFor := func(d time.Duration) func(context.Context) error {
		return func(ctx context.Context) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-timer.C:
				fmt.Println("timer:", d)
				return fmt.Errorf("abort!")
			case <-ctx.Done():
				fmt.Println("cancelled")
				return nil
			}
		}
	}
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.DoContext(ctx,
		sleepFor(1*time.Millisecond),
		sleepFor(10*time.Millisecond),
		sleepFor(100*time.Millisecond),
	)
	fmt.Println("err:", err)
	fmt.Println("duration:", time.Since(start).Round(time.Second))
	// Output:
	// timer: 1ms
	// cancelled
	// cancelled
	// err: abort!
	// duration: 0s
}
