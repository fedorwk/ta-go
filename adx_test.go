package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestADX(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	for _, period := range []int{2, 3, 14, 50} {
		got := apply(input, MustNewADX(period).Process)
		gotADX := make([]float64, len(got))
		gotPlus := make([]float64, len(got))
		gotMinus := make([]float64, len(got))
		for i, v := range got {
			gotADX[i], gotPlus[i], gotMinus[i] = v.ADX, v.PlusDI, v.MinusDI
		}
		lookback := 2*period - 1
		assertAlignedApprox(t, gotADX, talib.Adx(high, low, close, period), lookback, 1e-9)
		assertAlignedApprox(t, gotPlus, talib.PlusDI(high, low, close, period), lookback, 1e-9)
		assertAlignedApprox(t, gotMinus, talib.MinusDI(high, low, close, period), lookback, 1e-9)
	}
}
