package flowmatic

import (
	"context"
)

// DoAll runs fns in concurrently
// and waits for them all to finish.
// Each function receives a child context
// which is cancelled once one function returns an error or panics.
// DoAll returns nil if all functions succeed.
// Otherwise,
// DoAll returns a multierror containing the errors encountered.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func DoAll(ctx context.Context, fns ...func(context.Context) error) error {
	cg := WithContext(ctx)
	defer cg.Done()
	newfns := make([]func() error, 0, len(fns))
	for _, fn := range fns {
		newfns = append(newfns, cg.All(fn))
	}
	return Do(newfns...)
}
