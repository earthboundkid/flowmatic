package flowmatic

import (
	"context"
	"errors"
)

// Each starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and processes each item as a task.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func Each[Input any](numWorkers int, items []Input, task func(Input) error) error {
	return EachN(numWorkers, len(items), func(pos int) error {
		return task(items[pos])
	})
}

// EachN starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and starts a task for each number from 0 to numItems.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func EachN(numWorkers, numItems int, task func(int) error) error {
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

// EachMap starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and attempts to map the input slice to an output slice.
// Each task receives a child context.
// The first error or panic returned by a task
// cancels the child context
// and halts further task scheduling.
// The returned slice may contain partial results
// if an error is encountered in execution.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func EachMap[Input, Output any](numWorkers int, ctx context.Context, items []Input, task func(context.Context, Input) (Output, error)) (results []Output, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inch, ouch := TaskPool(numWorkers, func(pos int) (Output, error) {
		item := items[pos]
		return task(ctx, item)
	})

	var (
		panicVal any
		errs     []error
	)

	n := 0
	closeinch := false
	results = make([]Output, len(items))

	for {
		if n >= len(items) {
			closeinch = true
		}
		if closeinch && inch != nil {
			close(inch)
			inch = nil
		}
		select {
		case inch <- n:
			n++
		case r, ok := <-ouch:
			if !ok {
				if panicVal != nil {
					panic(panicVal)
				}
				return results, errors.Join(errs...)
			}
			if r.Err != nil {
				cancel()
				closeinch = true
				errs = append(errs, err)
			}
			if r.Panic != nil {
				cancel()
				closeinch = true
				panicVal = r.Panic
			}
			results[r.In] = r.Out
		}
	}
}
