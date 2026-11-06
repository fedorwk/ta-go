package ta

import "fmt"

func ExampleLag() {
	// The highest of the 20 values before the current one, as in a breakout
	// check that compares each value with the prior high.
	priorHigh := Chain(MustNewLag(1), MustNewMax(20))

	values := make([]float64, 20)
	for i := range values {
		values[i] = float64(i + 1) // 1..20
	}
	values[3] = 50 // the high of the first 20 values
	values = append(values, 45, 60, 55)

	for _, v := range values {
		if high, ok := priorHigh.Process(v); ok {
			fmt.Printf("value %.0f, prior high %.0f, breakout %v\n", v, high, v > high)
		}
	}

	// Output:
	// value 45, prior high 50, breakout false
	// value 60, prior high 50, breakout true
	// value 55, prior high 60, breakout false
}
