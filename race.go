package flowmatic

import (
	"context"
	"errors"
	"sync/atomic"
)

// DoContextRace runs fns in concurrently
// and waits for them all to finish.
// Each function receives a child context
// which is cancelled once one function has successfully completed or panicked.
// DoContextRace returns nil
// if at least one function completes without an error.
// If all functions return an error,
// DoContextRace returns a multierror containing all the errors.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func DoContextRace(ctx context.Context, fns ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make([]error, len(fns))
	var success atomic.Bool
	_ = eachN(len(fns), len(fns), func(pos int) error {
		defer func() {
			panicVal := recover()
			if panicVal != nil {
				cancel()
				panic(panicVal)
			}
		}()
		err := fns[pos](ctx)
		if err != nil {
			errs[pos] = err
			return nil
		}
		cancel()
		success.Store(true)
		return nil
	})
	if success.Load() {
		return nil
	}
	return errors.Join(errs...)
}
