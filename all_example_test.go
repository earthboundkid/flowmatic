package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/earthboundkid/flowmatic/v2"
)

func ExampleAll() {
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.All(ctx,
		func(ctx context.Context) error {
			// This task sleeps then returns an error
			d := 1 * time.Millisecond
			time.Sleep(d)
			fmt.Println("slept for", d)
			return fmt.Errorf("abort after %v", d)
		},
		func(ctx context.Context) error {
			// sleepFor is a cancelable time.Sleep.
			// The error of first task
			// causes the early cancelation of this one.
			if !sleepFor(ctx, 1*time.Minute) {
				fmt.Println("canceled")
			}
			return nil
		},
	)
	fmt.Println("err:", err)
	fmt.Println("exited early?", time.Since(start) < 10*time.Millisecond)
	// Output:
	// slept for 1ms
	// canceled
	// err: abort after 1ms
	// exited early? true
}
