package flowsafe_test

import (
	"fmt"

	"github.com/carlmjohnson/flowmatic"
	"github.com/carlmjohnson/flowmatic/flowsafe"
)

func ExampleMap() {
	var safemap flowsafe.Map[string, int]
	// Add to a map in concurrent goroutines
	err := flowmatic.Do(
		func() error {
			safemap.Add("a", 42)
			return nil
		},
		func() error {
			safemap.Add("b", 0x42)
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
	// Unwrap the map when done and use normally
	m := safemap.Unwrap()
	fmt.Println(m)
	// Output:
	// map[a:42 b:66]
}
