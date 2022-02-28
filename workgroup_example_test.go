package workgroup_test

import (
	"fmt"
	"math"

	"github.com/carlmjohnson/workgroup"
)

func Example() {
	in, out := workgroup.Start(1, func(i int) (float64, error) {
		if i < 0 {
			return 0, fmt.Errorf("imaginary: sqrt(%d)", i)
		}
		return math.Sqrt(float64(i)), nil
	})

	i := 10
loop:
	for {
		select {
		case in <- i:
			i--
			if i < -1 {
				close(in)
				in = nil
			}
		case r, ok := <-out:
			if !ok {
				break loop
			}
			if !r.Valid() {
				fmt.Printf("sqrt(%d) => %v\n", r.In, r.Err)
			} else {
				fmt.Printf("sqrt(%d) == %f\n", r.In, r.Out)
			}
		}
	}
	fmt.Println("done")

	// Output:
	// sqrt(10) == 3.162278
	// sqrt(9) == 3.000000
	// sqrt(8) == 2.828427
	// sqrt(7) == 2.645751
	// sqrt(6) == 2.449490
	// sqrt(5) == 2.236068
	// sqrt(4) == 2.000000
	// sqrt(3) == 1.732051
	// sqrt(2) == 1.414214
	// sqrt(1) == 1.000000
	// sqrt(0) == 0.000000
	// sqrt(-1) => imaginary: sqrt(-1)
	// done
}
