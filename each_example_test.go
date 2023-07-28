package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleEach() {
	times := []time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	}
	start := time.Now()
	err := flowmatic.Each(3, times, func(d time.Duration) error {
		time.Sleep(d)
		fmt.Println("slept", d)
		return nil
	})
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("executed concurrently?", time.Since(start) < 300*time.Millisecond)
	// Output:
	// slept 50ms
	// slept 100ms
	// slept 200ms
	// executed concurrently? true
}

func ExampleEach_cancel() {
	// To cancel execution early, communicate via a context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	times := []time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		300 * time.Millisecond,
	}
	task := func(d time.Duration) error {
		// simulate doing some work with a context
		t := time.NewTimer(d)
		defer t.Stop()

		select {
		case <-t.C:
			fmt.Println("slept", d)
		case <-ctx.Done():
			fmt.Println("canceled")
		}

		// if some condition applies, cancel the context for everyone
		if d == 100*time.Millisecond {
			cancel()
		}
		return nil
	}
	start := time.Now()
	if err := flowmatic.Each(3, times, task); err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("exited promptly?", time.Since(start) < 150*time.Millisecond)
	// Output:
	// slept 50ms
	// slept 100ms
	// canceled
	// exited promptly? true
}
