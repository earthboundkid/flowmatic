package flowmatic_test

import (
	"fmt"
	"time"

	"github.com/earthboundkid/flowmatic"
)

func ExampleDo() {
	start := time.Now()
	err := flowmatic.Do(
		func() error {
			time.Sleep(50 * time.Millisecond)
			fmt.Println("hello")
			return nil
		}, func() error {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("world")
			return nil
		}, func() error {
			time.Sleep(200 * time.Millisecond)
			fmt.Println("from flowmatic.Do")
			return nil
		})
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("executed concurrently?", time.Since(start) < 250*time.Millisecond)
	// Output:
	// hello
	// world
	// from flowmatic.Do
	// executed concurrently? true
}
