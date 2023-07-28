package flowmatic_test

import (
	"context"
	"fmt"
	"os"

	"github.com/carlmjohnson/flowmatic"
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
