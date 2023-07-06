package flowmatic

import (
	"context"
	"errors"
	"sync"
)

// DoRace runs fns in concurrently
// and waits for them all to finish.
// Each function receives a child context
// which is cancelled once one function has successfully completed or panicked.
// DoRace returns nil
// if at least one function completes without an error.
// If all functions return an error,
// DoRace returns a multierror containing all the errors.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func DoRace(ctx context.Context, fns ...func(context.Context) error) error {
	type result struct {
		err   error
		panic any
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errch := make(chan result, len(fns))

	wg.Add(len(fns))
	for i := range fns {
		fn := fns[i]
		go func() {
			defer wg.Done()
			defer func() {
				if panicVal := recover(); panicVal != nil {
					errch <- result{panic: panicVal}
				}
			}()
			errch <- result{err: fn(ctx)}
		}()
	}
	go func() {
		wg.Wait()
		close(errch)
	}()

	ok := false
	var panicVal any
	errs := make([]error, 0, len(fns))
	for res := range errch {
		switch {
		case res.err == nil && res.panic == nil:
			cancel()
			ok = true
		case res.panic != nil:
			cancel()
			panicVal = res.panic
		case res.err != nil:
			errs = append(errs, res.err)
		}
	}
	if panicVal != nil {
		panic(panicVal)
	}
	if ok {
		return nil
	}
	return errors.Join(errs...)
}
