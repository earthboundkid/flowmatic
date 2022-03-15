package workgroup_test

import (
	"fmt"
	"math"

	"github.com/carlmjohnson/workgroup"
)

func ExampleProcess() {
	var results []struct {
		int
		float64
	}
	workgroup.Process(1, func(i int) (float64, error) {
		if i < 0 {
			return 0, fmt.Errorf("imaginary: sqrt(%d)", i)
		}
		return math.Sqrt(float64(i)), nil
	}, func(i int, f float64, err error) ([]int, error) {
		results = append(results, struct {
			int
			float64
		}{i, f})
		if err == nil && math.Round(f) == f {
			return []int{int(f)}, nil
		}
		return nil, nil
	}, 256, 81)
	for _, r := range results {
		fmt.Printf("sqrt(%d) ~= %.2f\n", r.int, r.float64)
	}

	// Output:
	// sqrt(256) ~= 16.00
	// sqrt(81) ~= 9.00
	// sqrt(16) ~= 4.00
	// sqrt(9) ~= 3.00
	// sqrt(4) ~= 2.00
	// sqrt(3) ~= 1.73
	// sqrt(2) ~= 1.41
}
