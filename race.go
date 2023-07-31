package flowmatic

import (
	"context"
	"errors"
	"sync/atomic"
)

// Race runs each task concurrently
// and waits for them all to finish.
// Each function receives a child context
// which is cancelled once one function has successfully completed or panicked.
// Race returns nil
// if at least one function completes without an error.
// If all functions return an error,
// Race returns a multierror containing all the errors.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func Race(ctx context.Context, tasks ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make([]error, len(tasks))
	var success atomic.Bool
	_ = eachN(len(tasks), len(tasks), func(pos int) error {
		defer func() {
			panicVal := recover()
			if panicVal != nil {
				cancel()
				panic(panicVal)
			}
		}()
		err := tasks[pos](ctx)
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
