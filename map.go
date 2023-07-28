package flowmatic

import (
	"context"
)

// Map starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and attempts to map the input slice to an output slice.
// Each task receives a child context.
// The first error or panic returned by a task
// cancels the child context
// and halts further task scheduling.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func Map[Input, Output any](ctx context.Context, numWorkers int, items []Input, task func(context.Context, Input) (Output, error)) (results []Output, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inch, ouch := TaskPool(numWorkers, func(pos int) (Output, error) {
		item := items[pos]
		return task(ctx, item)
	})

	var panicVal any
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
				if err != nil {
					return nil, err
				}
				return results, nil
			}
			if r.Err != nil && err == nil {
				cancel()
				closeinch = true
				err = r.Err
			}
			if r.Panic != nil && panicVal == nil {
				cancel()
				closeinch = true
				panicVal = r.Panic
			}
			results[r.In] = r.Out
		}
	}
}
