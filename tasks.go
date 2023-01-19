package workgroup

import "errors"

type void = struct{}

// DoTasks starts n concurrent workers (or runtime.NumGoroutine workers if n < 1)
// and processes each input as a task.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
func DoTasks[Input any](n int, task func(Input) error, initial ...Input) error {
	errs := make([]error, 0, len(initial)+1)
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

// DoFuncs starts n concurrent workers (or runtime.NumGoroutine workers if n < 1)
// and executes each function in its own worker.
// Errors returned by a function do not halt execution,
// but are joined into a multierror return value.
func DoFuncs(n int, fns ...func() error) error {
	return DoTasks(n, func(in func() error) error {
		return in()
	}, fns...)
}
