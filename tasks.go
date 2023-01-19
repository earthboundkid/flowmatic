package workgroup

import "errors"

type void = struct{}

// DoTasks starts n concurrent workers (or GOMAXPROCS workers if n < 1)
// and processes each initial input as a task.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and returned as an error halting further execution.
func DoTasks[Input any](n int, task func(Input) error, initial ...Input) error {
	errs := make([]error, 0, len(initial))
	err := Do(n, func(in Input) (void, error) {
		return void{}, task(in)
	}, func(_ Input, _ void, err error) ([]Input, error) {
		if err != nil {
			errs = append(errs, err)
		}
		return nil, nil
	}, initial...)
	if err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// DoFuncs starts n concurrent workers (or GOMAXPROCS workers if n < 1)
// that execute each function.
// Errors returned by a function do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and returned as an error halting further execution.
func DoFuncs(n int, fns ...func() error) error {
	return DoTasks(n, func(in func() error) error {
		return in()
	}, fns...)
}
