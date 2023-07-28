package flowmatic_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
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

var (
	Web   = fakeSearch("web")
	Image = fakeSearch("image")
	Video = fakeSearch("video")
)

type Result string
type Search func(ctx context.Context, query string) (Result, error)

func fakeSearch(kind string) Search {
	return func(_ context.Context, query string) (Result, error) {
		return Result(fmt.Sprintf("%s result for %q", kind, query)), nil
	}
}

func Google(ctx context.Context, query string) ([]Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	searches := []Search{Web, Image, Video}
	results := make([]Result, len(searches))
	err := flowmatic.EachN(flowmatic.MaxProcs, len(searches),
		func(pos int) error {
			result, err := searches[pos](ctx, query)
			if err != nil {
				cancel()
				return err
			}
			results[pos] = result
			return nil
		})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func ExampleEachN_slice() {
	// Compare to https://pkg.go.dev/golang.org/x/sync/errgroup#example-Group-Parallel
	// and https://pkg.go.dev/sync#example-WaitGroup
	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// web result for "golang"
	// image result for "golang"
	// video result for "golang"
}

func ExampleEachN() {
	// Start with some slice of input work
	input := []string{"1", "42", "867-5309", "1337"}
	// Create a placeholder for output
	output := make([]int, len(input))
	// Concurrently process input and slot into output
	err := flowmatic.EachN(flowmatic.MaxProcs, len(input),
		func(pos int) error {
			n, err := strconv.Atoi(input[pos])
			if err != nil {
				return err
			}
			output[pos] = n
			return nil
		})
	if err != nil {
		// Couldn't process Jenny's number
		fmt.Println(err)
	}
	// Other values were processed
	fmt.Println(output)
	// Output:
	// strconv.Atoi: parsing "867-5309": invalid syntax
	// [1 42 0 1337]
}
