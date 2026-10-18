package ta

import (
	"errors"
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
		// go-talib computes variance as E[x^2]-E[x]^2, which is off by up to
		// ~1e-8 on this input; our bands are within ~1e-12 of a two-pass result.
		assertBands(t, got, input, upper, middle, lower, tc.period-1, 1e-7)
	}
}

func TestBollingerMA(t *testing.T) {
	input := randomWalk(1, 300)
	for _, mt := range allMATypes {
		for _, tc := range []struct {
			period int
			stddev float64
		}{
			{20, 2},
			{2, 1},
			{3, 2},
			{14, 2.5},
		} {
			got := apply(input, MustNewBollingerMA(tc.period, tc.stddev, mt).Process)
			var upper, middle, lower []float64
			if mt == MATypeRMA {
				// TA-Lib has no Wilder MA type; BBANDS is MA ± k*StdDev.
				middle = rmaReference(input, tc.period)
				sd := talib.StdDev(input, tc.period, tc.stddev)
				upper = make([]float64, len(input))
				lower = make([]float64, len(input))
				for i := tc.period - 1; i < len(input); i++ {
					upper[i], lower[i] = middle[i]+sd[i], middle[i]-sd[i]
				}
			} else {
				types := map[MAType]talib.MaType{MATypeSMA: talib.SMA, MATypeEMA: talib.EMA, MATypeWMA: talib.WMA}
				upper, middle, lower = talib.BBands(input, tc.period, tc.stddev, tc.stddev, types[mt])
			}
			assertBands(t, got, input, upper, middle, lower, tc.period-1, 1e-7)
		}
	}
}

func TestBollingerInvalid(t *testing.T) {
	for _, tc := range []struct {
		period int
		stddev float64
		mt     MAType
		want   error
	}{
		{0, 2, MATypeSMA, ErrInvalidPeriod},
		{20, 0, MATypeSMA, ErrInvalidStdDev},
		{20, 2, MAType(4), ErrInvalidMAType},
	} {
		if _, err := NewBollingerMA(tc.period, tc.stddev, tc.mt); !errors.Is(err, tc.want) {
			t.Errorf("NewBollingerMA(%d, %v, %d): got %v, want %v", tc.period, tc.stddev, tc.mt, err, tc.want)
		}
	}
}

// When the bands coincide PercentB is 0.5, and Bandwidth is 0 when the middle
// band is.
func TestBollingerFlat(t *testing.T) {
	for _, level := range []float64{0, 5} {
		bb := MustNewBollinger(3, 2)
		for i := 0; i < 5; i++ {
			got, ok := bb.Process(level)
			if !ok {
				continue
			}
			want := Bands{Upper: level, Middle: level, Lower: level, PercentB: 0.5}
			if got != want {
				t.Fatalf("level %v: got %+v, want %+v", level, got, want)
			}
		}
	}
}

// assertBands checks got against reference bands, zero-padded during warm-up
// as go-talib returns them, and checks PercentB and Bandwidth against those
// bands and the prices.
func assertBands(t *testing.T, got []Bands, prices, upper, middle, lower []float64, lookback int, tol float64) {
	t.Helper()
	bandwidth := make([]float64, len(prices))
	for i := lookback; i < len(prices); i++ {
		bandwidth[i] = (upper[i] - lower[i]) / middle[i]
	}
	// PercentB is ill-conditioned when the bands are close, so check that it
	// places the price within the reference bands, in price units: band
	// errors of up to tol move that point by up to tol*(1+2|PercentB|).
	t.Run("PercentB", func(t *testing.T) {
		for i, b := range got {
			j := i + lookback
			placed := lower[j] + b.PercentB*(upper[j]-lower[j])
			if math.Abs(placed-prices[j]) > tol*(1+2*math.Abs(b.PercentB)) {
				t.Fatalf("value mismatch at %d: PercentB %.8f places price at %.8f, want %.8f", i, b.PercentB, placed, prices[j])
			}
		}
	})
	for _, f := range []struct {
		name  string
		field func(Bands) float64
		want  []float64
	}{
		{"Upper", func(b Bands) float64 { return b.Upper }, upper},
		{"Middle", func(b Bands) float64 { return b.Middle }, middle},
		{"Lower", func(b Bands) float64 { return b.Lower }, lower},
		{"Bandwidth", func(b Bands) float64 { return b.Bandwidth }, bandwidth},
	} {
		values := make([]float64, len(got))
		for i, b := range got {
			values[i] = f.field(b)
		}
		t.Run(f.name, func(t *testing.T) {
			assertAlignedApprox(t, values, f.want, lookback, tol)
		})
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
