package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleRace() {
	task := func(d time.Duration) func(context.Context) error {
		return func(ctx context.Context) error {
			// sleepFor is a cancellable time.Sleep
			if !sleepFor(ctx, d) {
				fmt.Println("canceled")
				return ctx.Err()
			}
			fmt.Println("slept for", d)
			return nil
		}
	}
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.Race(ctx,
		task(1*time.Millisecond),
		task(10*time.Millisecond),
		task(100*time.Millisecond),
	)
	// Err is nil as long as one task succeeds
	fmt.Println("err:", err)
	fmt.Println("exited early?", time.Since(start) < 10*time.Millisecond)
	// Output:
	// slept for 1ms
	// canceled
	// canceled
	// err: <nil>
	// exited early? true
}

func ExampleRace_fakeRequest() {
	// Setup fake requests
	request := func(ctx context.Context, page string) (string, error) {
		var sleepLength time.Duration
		switch page {
		case "A":
			sleepLength = 10 * time.Millisecond
		case "B":
			sleepLength = 100 * time.Millisecond
		case "C":
			sleepLength = 10 * time.Second
		}
		if !sleepFor(ctx, sleepLength) {
			return "", ctx.Err()
		}
		return "got " + page, nil
	}
	ctx := context.Background()
	// Make variables to hold responses
	var pageA, pageB, pageC string
	// Race the requests to see who can answer first
	err := flowmatic.Race(ctx,
		func(ctx context.Context) error {
			var err error
			pageA, err = request(ctx, "A")
			return err
		},
		func(ctx context.Context) error {
			var err error
			pageB, err = request(ctx, "B")
			return err
		},
		func(ctx context.Context) error {
			var err error
			pageC, err = request(ctx, "C")
			return err
		},
	)
	fmt.Println("err:", err)
	fmt.Printf("A: %q B: %q C: %q\n", pageA, pageB, pageC)
	// Output:
	// err: <nil>
	// A: "got A" B: "" C: ""
}
