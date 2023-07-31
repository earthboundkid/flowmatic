package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleAll() {
	task := func(d time.Duration) func(context.Context) error {
		return func(ctx context.Context) error {
			// sleepFor is a cancellable time.Sleep
			if !sleepFor(ctx, d) {
				fmt.Println("cancelled")
				return nil
			}
			fmt.Println("slept for", d)
			return fmt.Errorf("abort after %v", d)
		}
	}
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.All(ctx,
		task(1*time.Millisecond),
		task(10*time.Millisecond),
		task(100*time.Millisecond),
	)
	fmt.Println("err:", err)
	fmt.Println("exited early?", time.Since(start) < 10*time.Millisecond)
	// Output:
	// slept for 1ms
	// cancelled
	// cancelled
	// err: abort after 1ms
	// exited early? true
}
