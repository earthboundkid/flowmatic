package flowmatic_test

import (
	"context"
	"fmt"
	"os"

	"github.com/carlmjohnson/flowmatic"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func fakeSearch(_ context.Context, kind, query string) (string, error) {
	return fmt.Sprintf("%s result for %q", kind, query), nil
}

func Google(ctx context.Context, query string) (map[string]string, error) {
	task := func(ctx context.Context, kind string) (string, error) {
		return fakeSearch(ctx, kind, query)
	}
	searches := []string{"web", "image", "video"}
	return flowmatic.Map(ctx, flowmatic.MaxProcs, searches, task)
}

func ExampleMap() {
	results, err := Google(context.Background(), "golang")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	keys := maps.Keys(results)
	slices.Sort(keys)
	for _, key := range keys {
		fmt.Println(results[key])
	}

	// Output:
	// image result for "golang"
	// video result for "golang"
	// web result for "golang"
}
