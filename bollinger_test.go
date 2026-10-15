package ta

import (
	"math"
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestBollinger(t *testing.T) {
	input := randomWalk(1, 300)
	for _, tc := range []struct {
		period int
		stddev float64
	}{
		{20, 2},
		{2, 1},
		{3, 2},
		{14, 2.5},
	} {
		got := apply(input, MustNewBollinger(tc.period, tc.stddev).Process)
		upper, middle, lower := talib.BBands(input, tc.period, tc.stddev, tc.stddev, talib.SMA)
		gotUpper := make([]float64, len(got))
		gotMiddle := make([]float64, len(got))
		gotLower := make([]float64, len(got))
		for i, v := range got {
			gotUpper[i], gotMiddle[i], gotLower[i] = v.Upper, v.Middle, v.Lower
		}
		// go-talib computes variance as E[x^2]-E[x]^2, which is off by up to
		// ~1e-8 on this input; our bands are within ~1e-12 of a two-pass result.
		assertAlignedApprox(t, gotUpper, upper, tc.period-1, 1e-7)
		assertAlignedApprox(t, gotMiddle, middle, tc.period-1, 1e-7)
		assertAlignedApprox(t, gotLower, lower, tc.period-1, 1e-7)
	}
}

// A small spread on a large price level must not be lost to cancellation:
// computing variance as E[x^2]-E[x]^2 gives a band width of ~1.57 here.
func TestBollingerHighPrice(t *testing.T) {
	bb := MustNewBollinger(20, 2)
	// Each window holds 1e8 + 0.001*{0..19}: population variance is
	// 0.001^2 * (20^2-1)/12.
	wantWidth := 2 * 0.001 * math.Sqrt(399.0/12)
	wantMiddle := 1e8 + 0.001*9.5
	for i := 0; i < 100000; i++ {
		got, ok := bb.Process(1e8 + 0.001*float64(i%20))
		if !ok {
			continue
		}
		if math.Abs(got.Middle-wantMiddle) > 1e-6 ||
			math.Abs(got.Upper-got.Middle-wantWidth) > 1e-6 ||
			math.Abs(got.Middle-got.Lower-wantWidth) > 1e-6 {
			t.Fatalf("input %d: got %+v, want middle %.6f and band width %.8f", i, got, wantMiddle, wantWidth)
		}
	}
}
