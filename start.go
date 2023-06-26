package flowmatic

import (
	"runtime"
	"sync"
)

// Result is the type returned by the output channel of TaskPool.
type Result[Input, Output any] struct {
	In    Input
	Out   Output
	Err   error
	Panic any
}

// TaskPool starts n workers (or GOMAXPROCS workers if n < 1) which consume
// the in channel, execute task, and send the Result on the out channel.
// Callers should close the in channel to stop the workers from waiting for tasks.
// The out channel will be closed once the last result has been sent.
func TaskPool[Input, Output any](n int, task Task[Input, Output]) (in chan<- Input, out <-chan Result[Input, Output]) {
	if n < 1 {
		n = runtime.GOMAXPROCS(0)
	}
	inch := make(chan Input)
	ouch := make(chan Result[Input, Output], n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			for inval := range inch {
				func() {
					defer func() {
						pval := recover()
						if pval == nil {
							return
						}
						ouch <- Result[Input, Output]{
							In:    inval,
							Panic: pval,
						}
					}()

					outval, err := task(inval)
					ouch <- Result[Input, Output]{inval, outval, err, nil}
				}()
			}
		}()
	}
	go func() {
		wg.Wait()
		close(ouch)
	}()
	return inch, ouch
}
