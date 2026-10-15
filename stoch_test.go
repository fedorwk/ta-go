package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestStoch(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	for _, tc := range []struct{ k, d int }{
		{14, 3},
		{5, 3},
		{3, 1},
		{1, 2},
	} {
		got := apply(input, MustNewStoch(tc.k, tc.d).Process)
		gotK := make([]float64, len(got))
		gotD := make([]float64, len(got))
		for i, v := range got {
			gotK[i], gotD[i] = v.K, v.D
		}
		// %K is passed through unsmoothed (slow-K period 1), matching Stoch.
		k, d := talib.Stoch(high, low, close, tc.k, 1, talib.SMA, tc.d, talib.SMA)
		lookback := tc.k - 1 + tc.d - 1
		assertAlignedApprox(t, gotK, k, lookback, 1e-9)
		assertAlignedApprox(t, gotD, d, lookback, 1e-9)
	}
}
