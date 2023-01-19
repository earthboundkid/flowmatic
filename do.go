package workgroup

import (
	"fmt"

	"github.com/carlmjohnson/deque"
)

// Use GOMAXPROCS workers when doing tasks.
const MaxProcs = -1

// Manager is a function that serially examines Task results to see if it produced any new Inputs.
type Manager[Input, Output any] func(Input, Output, error) ([]Input, error)

// Task is a function that can concurrently transform an input into an output.
type Task[Input, Output any] func(in Input) (out Output, err error)

// Do tasks using n concurrent workers (or GOMAXPROCS workers if n < 1)
// which produce output consumed by a serially run manager.
// The manager should return a slice of new task inputs based on prior task results,
// or return an error to halt processing.
// If a task panics during execution,
// the panic will be caught and returned as an error halting further execution.
func Do[Input, Output any](n int, task Task[Input, Output], manager Manager[Input, Output], initial ...Input) error {
	in, out := start(n, task)
	defer close(in)
	queue := deque.Of(initial...)
	inflight := 0
	for inflight > 0 || queue.Len() > 0 {
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
			if r.Panic != nil {
				return panicErr(r.Panic)
			}
			items, err := manager(r.In, r.Out, r.Err)
			if err != nil {
				return err
			}
			queue.Append(items...)
		}
	}
	return nil
}

func panicErr(v any) error {
	if e, ok := v.(error); ok {
		return e
	}
	return fmt.Errorf("panic: %v", v)
}
