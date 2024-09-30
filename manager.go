package flowmatic

import (
	"iter"

	"github.com/earthboundkid/deque/v2"
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

// Manage creates a Manager to run tasks concurrently
// using numWorkers concurrent workers (or GOMAXPROCS workers if numWorkers < 1).
func Manage[Input, Output any](numWorkers int, task Task[Input, Output]) *Manager[Input, Output] {
	return &Manager[Input, Output]{
		numWorkers: numWorkers,
		task:       task,
	}
}

// Manager is TKTK TODO
// which a sequence of TaskResults yielded serially.
// To add more jobs to call Queue on the Manager.
// If a task panics during execution,
// the panic will be caught and rethrown.
type Manager[Input, Output any] struct {
	numWorkers int
	task       Task[Input, Output]
	newItems   []Input
	in         Input
	out        Output
	err        error
	executing  bool
}

func (m *Manager[Input, Output]) Exec() func(func() bool) {
	if m.executing {
		panic("already executing")
	}
	return func(yield func() bool) {
		m.executing = true
		manager := func(in Input, out Output, err error) ([]Input, bool) {
			m.in, m.out, m.err = in, out, err
			m.newItems = m.newItems[:0]
			if !yield() {
				return nil, false
			}
			return m.newItems, true
		}

		manageTasks(m.numWorkers, m.task, manager, m.newItems...)
		m.executing = false
	}
}

func (m *Manager[Input, Output]) Queue(in ...Input) {
	m.newItems = append(m.newItems, in...)
}

func (m *Manager[Input, Output]) Input() Input {
	return m.in
}

func (m *Manager[Input, Output]) Output() Output {
	return m.out
}

func (m *Manager[Input, Output]) Error() error {
	return m.err
}

func (m *Manager[Input, Output]) Result() (Input, Output, error) {
	return m.in, m.out, m.err
}

func (m *Manager[Input, Output]) HasErr() bool {
	return m.err != nil
}

func (m *Manager[Input, Output]) Work() iter.Seq2[Input, Output] {
	return func(yield func(Input, Output) bool) {
		for range m.Exec() {
			if m.HasErr() || !yield(m.Input(), m.Output()) {
				return
			}
		}
	}
}
