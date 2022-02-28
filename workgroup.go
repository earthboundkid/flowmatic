package workgroup

import (
	"runtime"
	"sync"
)

// Result is the type returned by the output channel of Start.
type Result[Input, Output any] struct {
	In  Input
	Out Output
	Err error
}

// Valid reports whether Result.Err is nil.
func (r Result[Input, Output]) Valid() bool {
	return r.Err == nil
}

// Start n workers (or runtime.NumGoroutine workers if n < 1) which consume the in channel, execute task, and send the Result on the out channel. Callers should close the in channel to stop the workers from waiting for tasks. The out channel will be closed once the last result has been sent.
func Start[Input, Output any](n int, task func(in Input) (out Output, err error)) (in chan<- Input, out <-chan Result[Input, Output]) {
	if n < 1 {
		n = runtime.NumGoroutine()
	}
	inch := make(chan Input)
	ouch := make(chan Result[Input, Output], 1)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for inval := range inch {
				outval, err := task(inval)
				ouch <- Result[Input, Output]{inval, outval, err}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(ouch)
	}()
	return inch, ouch
}
