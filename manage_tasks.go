package flowmatic

import (
	"github.com/carlmjohnson/deque"
)

// Manager is a function that serially examines Task results to see if it produced any new Inputs.
// Returning false will halt the processing of future tasks.
type Manager[Input, Output any] func(Input, Output, error) (tasks []Input, ok bool)

// Task is a function that can concurrently transform an input into an output.
type Task[Input, Output any] func(in Input) (out Output, err error)

// ManageTasks manages tasks using numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// which produce output consumed by a serially run manager.
// The manager should return a slice of new task inputs based on prior task results,
// or return false to halt processing.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func ManageTasks[Input, Output any](numWorkers int, task Task[Input, Output], manager Manager[Input, Output], initial ...Input) {
	in, out := TaskPool(numWorkers, task)
	defer func() {
		close(in)
		// drain any waiting tasks
		for range out {
		}
	}()
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
			queue.RemoveFront()
		case r := <-out:
			inflight--
			if r.Panic != nil {
				panic(r.Panic)
			}
			items, ok := manager(r.In, r.Out, r.Err)
			if !ok {
				return
			}
			queue.PushBackSlice(items)
		}
	}
}
