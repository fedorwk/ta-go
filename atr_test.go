package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestATR(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	for _, period := range []int{1, 2, 3, 14, 50} {
		got := apply(input, MustNewATR(period).Process)
		assertAlignedApprox(t, got, talib.Atr(high, low, close, period), period, 1e-9)
	}
}
