package ta

import "fmt"

func ExampleChain() {
	candles := []OHLC{
		{High: 3, Low: 1, Close: 2},
		{High: 4, Low: 2, Close: 3},
		{High: 5, Low: 3, Close: 4},
		{High: 6, Low: 4, Close: 5},
	}
	// A 3-period moving average of the median price, (High+Low)/2.
	for _, avg := range Apply(Chain(HL2, MustNewSMA(3)), candles) {
		fmt.Printf("%.1f\n", avg)
	}

	// Output:
	// 3.0
	// 4.0
}
