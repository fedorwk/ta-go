package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestEMA(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{1, 2, 3, 14, 50} {
		got := apply(input, MustNewEMA(period).Process)
		assertAlignedApprox(t, got, talib.Ema(input, period), period-1, 1e-9)
	}
}
