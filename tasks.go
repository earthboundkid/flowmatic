package workgroup

import "errors"

type void = struct{}

// DoAll starts n concurrent workers (or GOMAXPROCS workers if n < 1)
// and processes each initial input as a task.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and rethrown in the main Goroutine.
func DoAll[Input any](n int, items []Input, task func(Input) error) error {
	errs := make([]error, 0, len(items))
	DoTasks(n, func(in Input) (void, error) {
		return void{}, task(in)
	}, func(_ Input, _ void, err error) ([]Input, bool) {
		if err != nil {
			errs = append(errs, err)
		}
		return nil, true
	}, items...)

	return errors.Join(errs...)
}

// DoFuncs starts n concurrent workers (or GOMAXPROCS workers if n < 1)
// that execute each function.
// Errors returned by a function do not halt execution,
// but are joined into a multierror return value.
// If a function panics during execution,
// the panic will be caught and rethrown in the main Goroutine.
func DoFuncs(n int, fns ...func() error) error {
	return DoAll(n, fns, func(in func() error) error {
		return in()
	})
}
