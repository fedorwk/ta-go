package ta

import "fmt"

func ExampleSMA() {
	s := MustNewSMA(3)
	for _, v := range []float64{1, 2, 3, 4, 5} {
		if avg, ok := s.Process(v); ok {
			fmt.Printf("%.1f\n", avg)
		}
	}

	// Output:
	// 2.0
	// 3.0
	// 4.0
}
