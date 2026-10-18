package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestDonchian(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	// go-talib's Max and Min return zeros for period 1.
	for _, period := range []int{2, 3, 20, 50} {
		got := apply(input, MustNewDonchian(period).Process)
		upper, lower := talib.Max(high, period), talib.Min(low, period)
		middle := make([]float64, len(upper))
		for i := range middle {
			middle[i] = (upper[i] + lower[i]) / 2
		}
		assertBands(t, got, close, upper, middle, lower, period-1, 0)
	}
}
