package flowmatic

import (
	"context"
)

// Map starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1),
// processes each item as a task,
// and creates a map from input to output.
// The first error returned by a task
// causes the context to be cancelled
// and stops further processing of items.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func Map[Input comparable, Output any](ctx context.Context, numWorkers int, items []Input, task func(context.Context, Input) (Output, error)) (map[Input]Output, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		out   Output
		panic any
	}

	runner := func(in Input) (r result, err error) {
		defer func() {
			r.panic = recover()
		}()
		r.out, err = task(ctx, in)
		return
	}
	var (
		recovered  any
		managerErr error
	)
	results := make(map[Input]Output, len(items))
	manager := func(in Input, r result, err error) ([]Input, bool) {
		if r.panic != nil {
			recovered = r.panic
			return nil, false
		} else if err != nil {
			if managerErr == nil {
				cancel()
				managerErr = err
			}
			return nil, false
		} else {
			results[in] = r.out
			return nil, true
		}
	}
	ManageTasks(numWorkers, runner, manager, items...)

	if recovered != nil {
		panic(recovered)
	}
	return results, managerErr
}
