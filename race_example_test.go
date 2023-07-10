package flowmatic_test

import (
	"context"
	"fmt"
	"time"

	"github.com/carlmjohnson/flowmatic"
)

func ExampleDoContextRace() {
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
	err := flowmatic.DoContextRace(ctx,
		sleepFor(1*time.Millisecond),
		sleepFor(1*time.Second),
		sleepFor(1*time.Minute),
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

func ExampleDoContextRace_fakeRequest() {
	// Cancellable sleep helper
	sleepFor := func(ctx context.Context, d time.Duration) bool {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-timer.C:
			return true
		case <-ctx.Done():
			return false
		}
	}

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
	err := flowmatic.DoContextRace(ctx,
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
