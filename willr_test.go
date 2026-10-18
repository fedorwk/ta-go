package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestWillR(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	for _, period := range []int{1, 2, 3, 14, 50} {
		got := apply(input, MustNewWillR(period).Process)
		assertAlignedApprox(t, got, talib.WillR(high, low, close, period), period-1, 1e-9)
	}
}
