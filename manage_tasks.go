package flowmatic

import (
	"iter"

	"github.com/carlmjohnson/deque"
)

// manager is a function that serially examines Task results to see if it produced any new Inputs.
// Returning false will halt the processing of future tasks.
type manager[Input, Output any] func(Input, Output, error) (tasks []Input, ok bool)

// Task is a function that can concurrently transform an input into an output.
type Task[Input, Output any] func(in Input) (out Output, err error)

// manageTasks manages tasks using numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1)
// which produce output consumed by a serially run manager.
// The manager should return a slice of new task inputs based on prior task results,
// or return false to halt processing.
// If a task panics during execution,
// the panic will be caught and rethrown in the parent Goroutine.
func manageTasks[Input, Output any](numWorkers int, task Task[Input, Output], manager manager[Input, Output], initial ...Input) {
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
		item, ok := queue.Front()
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

type TaskOutput[Input, Output any] struct {
	In       Input
	Out      Output
	Err      error
	pushtask func(Input)
}

func (to *TaskOutput[Input, Output]) HasErr() bool {
	return to.Err != nil
}

func (to *TaskOutput[Input, Output]) AddTask(in Input) {
	to.pushtask(in)
}

func Tasks[Input, Output any](numWorkers int, task Task[Input, Output], initial ...Input) iter.Seq[*TaskOutput[Input, Output]] {
	return func(yield func(*TaskOutput[Input, Output]) bool) {
		manager := func(in Input, out Output, err error) ([]Input, bool) {
			var newitems []Input
			to := TaskOutput[Input, Output]{
				In:  in,
				Out: out,
				Err: err,
				pushtask: func(newin Input) {
					newitems = append(newitems, newin)
				},
			}
			if !yield(&to) {
				return nil, false
			}
			return newitems, true
		}

		manageTasks(numWorkers, task, manager, initial...)
	}
}
