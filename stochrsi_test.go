package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestStochRSI(t *testing.T) {
	input := randomWalk(1, 300)
	for _, tc := range []struct{ rsi, k, d int }{
		{14, 14, 3},
		{14, 5, 3},
		{5, 3, 1},
		{2, 1, 2},
	} {
		got := apply(input, MustNewStochRSI(tc.rsi, tc.k, tc.d).Process)
		gotK := make([]float64, len(got))
		gotD := make([]float64, len(got))
		for i, v := range got {
			gotK[i], gotD[i] = v.K, v.D
		}
		k, d := talib.StochRsi(input, tc.rsi, tc.k, tc.d, talib.SMA)
		lookback := tc.rsi + tc.k - 1 + tc.d - 1
		assertAlignedApprox(t, gotK, k, lookback, 1e-9)
		assertAlignedApprox(t, gotD, d, lookback, 1e-9)
	}
}
