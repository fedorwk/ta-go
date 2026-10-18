package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestStdDev(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{2, 3, 14, 50} {
		got := apply(input, MustNewStdDev(period).Process)
		// go-talib computes variance as E[x^2]-E[x]^2; see TestBollinger.
		assertAlignedApprox(t, got, talib.StdDev(input, period, 1), period-1, 1e-7)
	}
}
