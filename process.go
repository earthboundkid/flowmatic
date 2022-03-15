package workgroup

import "github.com/carlmjohnson/deque"

// Manager is a function that serially examines Task results to see if it produced any new Inputs.
type Manager[Input, Output any] func(Input, Output, error) ([]Input, error)

// Process tasks using n concurrent workers (or runtime.NumGoroutine workers if n < 1)
// which produce output consumed by a serially run manager. The manager should return a slice of
// new task inputs based on prior task results, or return an error to halt processing.
func Process[Input, Output any](n int, task Task[Input, Output], manager Manager[Input, Output], initial ...Input) error {
	in, out := Start[Input, Output](n, task)
	defer close(in)
	queue := deque.Of(initial...)
	inflight := 0
	for {
		inch := in
		item, ok := queue.Head()
		if !ok {
			inch = nil
		}
		select {
		case inch <- item:
			inflight++
			queue.PopHead()
		case r := <-out:
			inflight--
			items, err := manager(r.In, r.Out, r.Err)
			if err != nil {
				return err
			}
			queue.Append(items...)
		}
		if inflight == 0 && queue.Len() == 0 {
			return nil
		}
	}
}
