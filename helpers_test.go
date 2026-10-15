package ta

import (
	"math"
	"math/rand"
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

// assertAlignedApprox checks that got starts at input index lookback of a
// go-talib-style reference (want has one entry per input, zero-padded during
// warm-up) and matches it from there on.
func assertAlignedApprox(t *testing.T, got, want []float64, lookback int, tol float64) {
	t.Helper()
	if len(got) != len(want)-lookback {
		t.Fatalf("first output at input index %d, want %d", len(want)-len(got), lookback)
	}
	assertFloatSliceApprox(t, got, want[lookback:], tol)
}

func randomWalk(seed int64, n int) []float64 {
	r := rand.New(rand.NewSource(seed))
	out := make([]float64, n)
	p := 100.0
	for i := range out {
		p += r.NormFloat64()
		out[i] = p
	}
	return out
}

// randomCandles builds n candles whose closes follow a random walk; each
// opens at the previous close and has high/low beyond the open-close range.
func randomCandles(seed int64, n int) []OHLCV {
	r := rand.New(rand.NewSource(seed))
	out := make([]OHLCV, n)
	prev := 100.0
	for i := range out {
		c := prev + r.NormFloat64()
		out[i] = OHLCV{
			Open:   prev,
			High:   math.Max(prev, c) + r.ExpFloat64()*0.5,
			Low:    math.Min(prev, c) - r.ExpFloat64()*0.5,
			Close:  c,
			Volume: 100 + r.Float64()*1000,
		}
		prev = c
	}
	return out
}

func randomOHLC(seed int64, n int) []OHLC {
	candles := randomCandles(seed, n)
	out := make([]OHLC, n)
	for i, c := range candles {
		out[i] = OHLC{Open: c.Open, High: c.High, Low: c.Low, Close: c.Close}
	}
	return out
}
