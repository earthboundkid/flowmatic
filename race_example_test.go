package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

// sleepFor is a cancellable sleep
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

func ExampleRace() {
	task := func(d time.Duration) func(context.Context) error {
		return func(ctx context.Context) error {
			if sleepFor(ctx, d) {
				fmt.Println("timer:", d)
				return nil
			}
			fmt.Println("cancelled")
			return ctx.Err()
		}
	}
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.Race(ctx,
		task(1*time.Millisecond),
		task(1*time.Second),
		task(1*time.Minute),
	)
	fmt.Println("err:", err)
	fmt.Println("duration:", time.Since(start).Round(time.Second))
	// Output:
	// timer: 1ms
	// cancelled
	// cancelled
	// err: <nil>
	// duration: 0s
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
