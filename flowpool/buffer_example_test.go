package flowpool_test

import (
	"fmt"
	"testing"

	"github.com/carlmjohnson/flowmatic/flowpool"
)

func ExampleGetBuffer() {
	f := func() {
		buf := flowpool.GetBuffer()
		defer flowpool.PutBuffer(buf)

		buf.WriteString("Hello, World!")
	}

	// run once to prime the pool
	f()

	allocs := testing.AllocsPerRun(100, f)
	fmt.Println("allocs:", allocs)
	// Output:
	// allocs: 0
}
