package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestCCI(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	for _, period := range []int{2, 3, 14, 20, 50} {
		got := apply(input, MustNewCCI(period).Process)
		assertAlignedApprox(t, got, talib.Cci(high, low, close, period), period-1, 1e-9)
	}
}
