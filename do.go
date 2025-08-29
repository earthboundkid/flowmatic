package flowmatic

import (
	"errors"
	"sync"
)

// Do runs each task concurrently
// and waits for them all to finish.
// Errors returned by tasks do not cancel execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func Do(tasks ...func() error) error {
	type result struct {
		err   error
		panic any
	}

	var wg sync.WaitGroup
	errch := make(chan result, len(tasks))

	for _, fn := range tasks {
		wg.Go(func() {
			defer func() {
				if panicVal := recover(); panicVal != nil {
					errch <- result{panic: panicVal}
				}
			}()
			errch <- result{err: fn()}
		})
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
