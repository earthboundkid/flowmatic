package flowsafe_test

import (
	"fmt"
	"strconv"

	"github.com/carlmjohnson/flowmatic"
	"github.com/carlmjohnson/flowmatic/flowsafe"
)

func ExampleCount() {
	// Start with some slice of input work
	input := []string{"1", "42", "867-5309", "1337"}
	// Create a holder for output
	output := make([]int, len(input))
	// Concurrently process input and slot into output
	err := flowmatic.Each(flowmatic.MaxProcs, flowsafe.Count(input),
		func(item flowsafe.Enum[string]) error {
			n, err := strconv.Atoi(*item.Value)
			if err != nil {
				return err
			}
			output[item.Pos] = n
			return nil
		})
	if err != nil {
		// Couldn't process Jenny's number
		fmt.Println(err)
	}
	// Other values were processed
	fmt.Println(output)
	// Output:
	// strconv.Atoi: parsing "867-5309": invalid syntax
	// [1 42 0 1337]
}
