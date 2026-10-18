package ta

import (
	"errors"
	"fmt"
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
		gotK, gotD := splitStoch(got)
		// %K is passed through unsmoothed (slow-K period 1), matching Stoch.
		k, d := talib.Stoch(high, low, close, tc.k, 1, talib.SMA, tc.d, talib.SMA)
		lookback := tc.k - 1 + tc.d - 1
		assertAlignedApprox(t, gotK, k, lookback, 1e-9)
		assertAlignedApprox(t, gotD, d, lookback, 1e-9)
	}
}

func TestStochFull(t *testing.T) {
	input := randomOHLC(1, 300)
	for _, mt := range allMATypes {
		for _, tc := range []struct{ k, slowK, d int }{
			{14, 3, 3},
			{5, 3, 3},
			{14, 1, 3},
			{14, 3, 1},
			{1, 2, 2},
			{9, 5, 4},
		} {
			t.Run(fmt.Sprintf("%v/%d,%d,%d", mt, tc.k, tc.slowK, tc.d), func(t *testing.T) {
				got := apply(input, MustNewStochFull(tc.k, tc.slowK, tc.d, mt).Process)
				gotK, gotD := splitStoch(got)
				k, d := talibStoch(input, tc.k, tc.slowK, tc.d, mt)
				lookback := tc.k - 1 + tc.slowK - 1 + tc.d - 1
				assertAlignedApprox(t, gotK, k, lookback, 1e-9)
				assertAlignedApprox(t, gotD, d, lookback, 1e-9)
			})
		}
	}
}

func TestStochFullInvalid(t *testing.T) {
	for _, tc := range []struct {
		k, slowK, d int
		mt          MAType
		want        error
	}{
		{0, 3, 3, MATypeSMA, ErrInvalidPeriod},
		{14, 0, 3, MATypeSMA, ErrInvalidPeriod},
		{14, 3, 0, MATypeSMA, ErrInvalidPeriod},
		{14, 3, 3, MAType(-1), ErrInvalidMAType},
		{14, 3, 3, MAType(4), ErrInvalidMAType},
	} {
		if _, err := NewStochFull(tc.k, tc.slowK, tc.d, tc.mt); !errors.Is(err, tc.want) {
			t.Errorf("NewStochFull(%d, %d, %d, %d): got %v, want %v", tc.k, tc.slowK, tc.d, tc.mt, err, tc.want)
		}
	}
}

// talibStoch returns TA-Lib's STOCH with both moving averages of type mt.
//
// go-talib's Stoch computes %D over the zero-padded slow-%K buffer, which
// corrupts an EMA's seed when slowK > 1. TA-Lib C averages only the defined
// values, so in that case, and for RMA, which TA-Lib lacks, both lines are
// built the same way from go-talib's fast %K.
func talibStoch(input []OHLC, k, slowK, d int, mt MAType) ([]float64, []float64) {
	high, low, close := ohlcToHighLowClose(input)
	if mt != MATypeRMA && (mt != MATypeEMA || slowK == 1) {
		types := map[MAType]talib.MaType{MATypeSMA: talib.SMA, MATypeEMA: talib.EMA, MATypeWMA: talib.WMA}
		return talib.Stoch(high, low, close, k, slowK, types[mt], d, types[mt])
	}
	fastK, _ := talib.Stoch(high, low, close, k, 1, talib.SMA, 1, talib.SMA)
	slowKs := padFront(talibMA(fastK[k-1:], slowK, mt), k-1)
	slowKLookback := k - 1 + slowK - 1
	ds := padFront(talibMA(slowKs[slowKLookback:], d, mt), slowKLookback)
	return slowKs, ds
}

// padFront prepends n zeros to values.
func padFront(values []float64, n int) []float64 {
	return append(make([]float64, n), values...)
}

func splitStoch(values []StochValue) ([]float64, []float64) {
	k := make([]float64, len(values))
	d := make([]float64, len(values))
	for i, v := range values {
		k[i], d[i] = v.K, v.D
	}
	return k, d
}
