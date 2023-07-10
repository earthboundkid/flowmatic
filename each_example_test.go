package flowmatic_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/carlmjohnson/flowmatic"
	"github.com/carlmjohnson/flowmatic/flowsafe"
	"golang.org/x/exp/slices"
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

func fakeSearch(_ context.Context, kind, query string) (string, error) {
	return fmt.Sprintf("%s result for %q", kind, query), nil
}

func Google(ctx context.Context, query string) ([]string, error) {
	searches := []string{"web", "image", "video"}
	results := flowsafe.MakeSlice[string](len(searches))

	task := func(kind string) error {
		result, err := fakeSearch(ctx, kind, query)
		if err != nil {
			return err
		}
		results.Push(result)
		return nil
	}

	err := flowmatic.Each(flowmatic.MaxProcs, searches, task)
	if err != nil {
		return nil, err
	}
	r := results.Unwrap()
	slices.Sort(r)
	return r, nil
}

func ExampleEach_slice() {
	// Compare to https://pkg.go.dev/sync#example-WaitGroup
	// and https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Parallel
	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// image result for "golang"
	// video result for "golang"
	// web result for "golang"
}
