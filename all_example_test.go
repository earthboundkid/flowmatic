package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleDoContext() {
	task := func(d time.Duration) func(context.Context) error {
		return func(ctx context.Context) error {
			if sleepFor(ctx, d) {
				fmt.Println("timer:", d)
				return fmt.Errorf("abort!")
			}
			fmt.Println("cancelled")
			return nil
		}
	}
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.DoContext(ctx,
		task(1*time.Millisecond),
		task(10*time.Millisecond),
		task(100*time.Millisecond),
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
