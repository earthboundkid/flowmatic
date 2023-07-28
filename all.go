package flowmatic

import (
	"context"
)

// DoContext runs fns in concurrently
// and waits for them all to finish.
// Each function receives a child context
// which is cancelled once one function returns an error or panics.
// DoContext returns nil if all functions succeed.
// Otherwise,
// DoContext returns a multierror containing the errors encountered.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func DoContext(ctx context.Context, fns ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return eachN(len(fns), len(fns), func(pos int) error {
		defer func() {
			panicVal := recover()
			if panicVal != nil {
				cancel()
				panic(panicVal)
			}
		}()
		err := fns[pos](ctx)
		if err != nil {
			cancel()
			return err
		}
		return nil
	})
}
