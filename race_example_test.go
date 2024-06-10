package flowmatic_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/earthboundkid/flowmatic"
)

func ExampleRace() {
	ctx := context.Background()
	start := time.Now()
	err := flowmatic.Race(ctx,
		func(ctx context.Context) error {
			// This task sleeps for only 1ms
			d := 1 * time.Millisecond
			time.Sleep(d)
			fmt.Println("slept for", d)
			return nil
		},
		func(ctx context.Context) error {
			// This task wants to sleep for a whole minute.
			d := 1 * time.Minute
			// But sleepFor is a cancelable time.Sleep.
			// So when the other task completes,
			// it cancels this one, causing it to return early.
			if !sleepFor(ctx, d) {
				fmt.Println("canceled")
			}
			// The error here is ignored
			// because the other task succeeded
			return errors.New("ignored")
		},
	)
	// Err is nil as long as one task succeeds
	fmt.Println("err:", err)
	fmt.Println("exited early?", time.Since(start) < 10*time.Millisecond)
	// Output:
	// slept for 1ms
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
