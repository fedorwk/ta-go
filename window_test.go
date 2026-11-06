package ta

import (
	"errors"
	"math"
	"testing"

	talib "github.com/markcheno/go-talib"
)

var windowPeriods = []int{1, 2, 14, 50}

func TestMaxMin(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range windowPeriods {
		wantMax, wantMin := talib.Max(input, period), talib.Min(input, period)
		if period == 1 {
			// go-talib's Max and Min return zeros for period 1, where the
			// extremum of a one-value window is the value itself.
			wantMax, wantMin = input, input
		}
		assertAlignedApprox(t, apply(input, MustNewMax(period).Process), wantMax, period-1, 0)
		assertAlignedApprox(t, apply(input, MustNewMin(period).Process), wantMin, period-1, 0)
	}
}

func TestMaxMinFlat(t *testing.T) {
	input := []float64{0, 0, 0, 0, 0}
	assertFloatSliceApprox(t, apply(input, MustNewMax(3).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply(input, MustNewMin(3).Process), []float64{0, 0, 0}, 0)
}

func TestSum(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range windowPeriods {
		got := apply(input, MustNewSum(period).Process)
		assertAlignedApprox(t, got, talib.Sum(input, period), period-1, 1e-9)
	}
}

func TestSumZeros(t *testing.T) {
	input := []float64{0, 0, 0, 0}
	assertFloatSliceApprox(t, apply(input, MustNewSum(2).Process), []float64{0, 0, 0}, 0)
}

// TestSumLargeOffset checks that the running sum stays within a few rounding
// errors of a from-scratch sum over a long stream of large values, where an
// incrementally updated sum drifts.
func TestSumLargeOffset(t *testing.T) {
	input := randomWalk(2, 100_000)
	for i := range input {
		input[i] += 1e5
	}
	for _, period := range []int{2, 14, 50, 390} {
		got := apply(input, MustNewSum(period).Process)
		if len(got) != len(input)-period+1 {
			t.Fatalf("period %d: got %d outputs, want %d", period, len(got), len(input)-period+1)
		}
		for i, g := range got {
			want := 0.0
			for _, x := range input[i : i+period] {
				want += x
			}
			// Between exact recomputes the running sum takes at most period
			// increments, each rounding by one ulp of the sum.
			tol := 4 * float64(period) * math.Abs(want) * 0x1p-52
			if math.Abs(g-want) > tol {
				t.Fatalf("period %d output %d: got %.10f, want %.10f (diff %g > %g)", period, i, g, want, g-want, tol)
			}
		}
	}
}

func TestLag(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{1, 2, 3, 14, 50} {
		got := apply(input, MustNewLag(period).Process)
		want := make([]float64, len(input))
		for i := period; i < len(input); i++ {
			want[i] = input[i-period]
		}
		assertAlignedApprox(t, got, want, period, 0)
	}
}

func TestTrueRange(t *testing.T) {
	input := randomOHLC(1, 300)
	high, low, close := ohlcToHighLowClose(input)
	got := apply(input, MustNewTrueRange().Process)
	assertAlignedApprox(t, got, talib.TRange(high, low, close), 1, 0)
}

func TestTrueRangeEdges(t *testing.T) {
	input := []OHLC{
		{Open: 10, High: 12, Low: 9, Close: 10},
		{Open: 10, High: 10, Low: 10, Close: 10}, // flat at the previous close
		{Open: 13, High: 14, Low: 13, Close: 14}, // gap up: |L-Cprev| = 3 > 1
		{Open: 8, High: 9, Low: 7, Close: 8},     // gap down: |L-Cprev| = 7
	}
	got := apply(input, MustNewTrueRange().Process)
	assertFloatSliceApprox(t, got, []float64{0, 4, 7}, 0)
}

func TestWindowPrimitivesInvalidPeriod(t *testing.T) {
	constructors := map[string]func(int) error{
		"Max": func(p int) error { _, err := NewMax(p); return err },
		"Min": func(p int) error { _, err := NewMin(p); return err },
		"Sum": func(p int) error { _, err := NewSum(p); return err },
		"Lag": func(p int) error { _, err := NewLag(p); return err },
	}
	for name, newX := range constructors {
		for _, p := range []int{0, -1} {
			if err := newX(p); !errors.Is(err, ErrInvalidPeriod) {
				t.Errorf("New%s(%d): got %v, want ErrInvalidPeriod", name, p, err)
			}
		}
	}
	mustConstructors := map[string]func(){
		"Max": func() { MustNewMax(0) },
		"Min": func() { MustNewMin(0) },
		"Sum": func() { MustNewSum(0) },
		"Lag": func() { MustNewLag(0) },
	}
	for name, mustNewX := range mustConstructors {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("MustNew%s(0) did not panic", name)
				}
			}()
			mustNewX()
		}()
	}
}
