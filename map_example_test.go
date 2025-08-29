package flowmatic_test

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/earthboundkid/flowmatic/v2"
)

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
	searches := []Search{Web, Image, Video}
	return flowmatic.Map(ctx, flowmatic.MaxProcs, searches,
		func(ctx context.Context, search Search) (Result, error) {
			return search(ctx, query)
		})
}

func ExampleMap() {
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

func ExampleMap_simple() {
	ctx := context.Background()

	// Start with some slice of input work
	input := []string{"0", "1", "42", "1337"}
	// Have a task that takes a context
	decodeAndDouble := func(ctx context.Context, s string) (int, error) {
		// Do some work
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		// Return early if context was canceled
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}
		// Do more work
		return 2 * n, nil
	}
	// Concurrently process input into output
	output, err := flowmatic.Map(ctx, flowmatic.MaxProcs, input, decodeAndDouble)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(output)
	// Output:
	// [0 2 84 2674]
}
