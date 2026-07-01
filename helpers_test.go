package ta

import (
	"math"
	"testing"
)

func apply[I any, O any](input []I, update func(I) (O, bool)) []O {
	var out []O
	for _, v := range input {
		if o, ok := update(v); ok {
			out = append(out, o)
		}
	}
	return out
}

func assertFloatSliceApprox(t *testing.T, got, want []float64, tol float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d want %d", len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i]-want[i]) > tol {
			t.Fatalf("value mismatch at %d: got %.8f want %.8f", i, got[i], want[i])
		}
	}
}

func ohlcToHighLowClose(values []OHLC) ([]float64, []float64, []float64) {
	high := make([]float64, len(values))
	low := make([]float64, len(values))
	close := make([]float64, len(values))
	for i, v := range values {
		high[i] = v.High
		low[i] = v.Low
		close[i] = v.Close
	}
	return high, low, close
}

func tail(values []float64, n int) []float64 {
	if n <= 0 {
		return nil
	}
	if n >= len(values) {
		return values
	}
	return values[len(values)-n:]
}