package flowmatic

import (
	"context"
)

// All runs each task concurrently
// and waits for them all to finish.
// Each task receives a child context
// which is cancelled once one task returns an error or panics.
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
