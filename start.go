package workgroup

import (
	"runtime"
	"sync"
)

// result is the type returned by the output channel of Start.
type result[Input, Output any] struct {
	In    Input
	Out   Output
	Err   error
	Panic any
}

// start n workers (or runtime.NumGoroutine workers if n < 1) which consume
// the in channel, execute task, and send the Result on the out channel.
// Callers should close the in channel to stop the workers from waiting for tasks.
// The out channel will be closed once the last result has been sent.
func start[Input, Output any](n int, task Task[Input, Output]) (in chan<- Input, out <-chan result[Input, Output]) {
	if n < 1 {
		n = runtime.NumGoroutine()
	}
	inch := make(chan Input)
	ouch := make(chan result[Input, Output], n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			defer func() {
				pval := recover()
				if pval == nil {
					return
				}
				ouch <- result[Input, Output]{Panic: pval}
			}()
			for inval := range inch {
				outval, err := task(inval)
				ouch <- result[Input, Output]{inval, outval, err, nil}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(ouch)
	}()
	return inch, ouch
}
