package workgroup

// Do starts each function in its own goroutine.
// Errors returned by a function do not halt execution,
// but are joined into a multierror return value.
// If a function panics during execution,
// a panic will be caught and rethrown in the parent Goroutine.
func Do(fns ...func() error) error {
	return DoEach(len(fns), fns, func(in func() error) error {
		return in()
	})
}
