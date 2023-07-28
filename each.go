package flowmatic

import (
	"errors"
)

// Each starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and processes each item as a task.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func Each[Input any](numWorkers int, items []Input, task func(Input) error) error {
	return eachN(numWorkers, len(items), func(pos int) error {
		return task(items[pos])
	})
}

// eachN starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and starts a task for each number from 0 to numItems.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func eachN(numWorkers, numItems int, task func(int) error) error {
	type void struct{}
	inch, ouch := TaskPool(numWorkers, func(pos int) (void, error) {
		return void{}, task(pos)
	})
	var (
		panicVal any
		errs     []error
	)
	_ = Do(
		func() error {
			for i := 0; i < numItems; i++ {
				inch <- i
			}
			close(inch)
			return nil
		},
		func() error {
			for r := range ouch {
				if r.Panic != nil {
					panicVal = r.Panic
				}
				if r.Err != nil {
					errs = append(errs, r.Err)
				}
			}
			return nil
		})
	if panicVal != nil {
		panic(panicVal)
	}
	return errors.Join(errs...)
}
