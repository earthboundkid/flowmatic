package workgroup

// Do starts n concurrent workers (or GOMAXPROCS workers if n < 1)
// that execute each function.
// Errors returned by a function do not halt execution,
// but are joined into a multierror return value.
// If a function panics during execution,
// the panic will be caught and rethrown in the main Goroutine.
func Do(n int, fns ...func() error) error {
	return DoEach(n, fns, func(in func() error) error {
		return in()
	})
}
