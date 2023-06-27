package flowmatic

import "errors"

// Each starts numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// and processes each item as a task.
// Errors returned by a task do not halt execution,
// but are joined into a multierror return value.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func Each[Input any](numWorkers int, items []Input, task func(Input) error) error {
	var recovered any
	errs := make([]error, 0, len(items))
	runner := func(in Input) (r any, err error) {
		defer func() {
			r = recover()
		}()
		err = task(in)
		return
	}
	manager := func(_ Input, r any, err error) ([]Input, bool) {
		if r != nil {
			recovered = r
		}
		if err != nil {
			errs = append(errs, err)
		}
		return nil, true
	}
	ManageTasks(numWorkers, runner, manager, items...)

	if recovered != nil {
		panic(recovered)
	}
	return errors.Join(errs...)
}
