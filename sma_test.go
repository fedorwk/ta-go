package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestSMA(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{1, 2, 3, 14, 50} {
		got := apply(input, MustNewSMA(period).Process)
		assertAlignedApprox(t, got, talib.Sma(input, period), period-1, 1e-9)
	}
}
