package flowmatic_test

import (
	"fmt"
	"slices"
	"time"

	"github.com/earthboundkid/flowmatic/v2"
)

func ExampleEach() {
	times := slices.Values([]time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	})
	start := time.Now()
	err := flowmatic.Each(3, times, func(d time.Duration) error {
		time.Sleep(d)
		fmt.Println("slept", d)
		return nil
	})
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println("executed concurrently?", time.Since(start) < 300*time.Millisecond)
	// Output:
	// slept 50ms
	// slept 100ms
	// slept 200ms
	// executed concurrently? true
}
