package flowmatic

import (
	"context"
	"errors"
)

// All runs each task concurrently
// and waits for them all to finish.
// Each task receives a child context
// which is canceled once one task returns an error or panics.
// All returns nil if all tasks succeed.
// Otherwise,
// All returns a multierror containing the errors encountered.
// If a task panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func All(ctx context.Context, tasks ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	return eachN(len(tasks), len(tasks), func(pos int) error {
		defer func() {
			panicVal := recover()
			if panicVal != nil {
				cancel()
				panic(panicVal)
			}
		}()
		err := tasks[pos](ctx)
		if err != nil {
			cancel()
			return err
		}
		return nil
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
				if r.Panic != nil && panicVal == nil {
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
