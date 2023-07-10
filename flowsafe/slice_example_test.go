package flowsafe_test

import (
	"fmt"

	"github.com/carlmjohnson/flowmatic"
	"github.com/carlmjohnson/flowmatic/flowsafe"
	"golang.org/x/exp/slices"
)

func ExampleSlice() {
	var safeslice flowsafe.Slice[string]
	// Push items to the slice in concurrent goroutines
	err := flowmatic.Do(
		func() error {
			safeslice.Push("a")
			return nil
		},
		func() error {
			safeslice.Push("b")
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
	// Unwrap the slice when done and use normally
	s := safeslice.Unwrap()
	slices.Sort(s)
	fmt.Println(s)
	// Output:
	// [a b]
}
