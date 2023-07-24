package flowmatic

import (
	"errors"
	"sync"
)

// Do starts each function in its own goroutine.
// Errors returned by a function do not halt execution,
// but are joined into a multierror return value.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func Do(fns ...func() error) error {
	type result struct {
		err   error
		panic any
	}

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
			errch <- result{err: fn()}
		}()
	}
	go func() {
		wg.Wait()
		close(errch)
	}()

	var (
		panicVal any
		errs     []error
	)
	for res := range errch {
		switch {
		case res.err == nil && res.panic == nil:
			continue
		case res.panic != nil:
			panicVal = res.panic
		case res.err != nil:
			errs = append(errs, res.err)
		}
	}
	if panicVal != nil {
		panic(panicVal)
	}
	return errors.Join(errs...)
}
