package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestRSI(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{2, 3, 14, 50} {
		got := apply(input, MustNewRSI(period).Process)
		assertAlignedApprox(t, got, talib.Rsi(input, period), period, 1e-9)
	}
}
