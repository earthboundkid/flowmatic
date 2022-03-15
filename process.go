package workgroup

import "github.com/carlmjohnson/deque"

func Process[Input, Output any](n int, task func(in Input) (Output, error), manager func(Input, Output, error) ([]Input, error), initial ...Input) error {
	in, out := Start[Input, Output](1, task)
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
