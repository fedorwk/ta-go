package ta

import (
	"math"
	"testing"
)

// On a linear ramp with slope 1, an EMA seeded with a simple average lags the
// input by exactly (period-1)/2, so the MACD line is the constant
// (slow-fast)/2 and the signal line equals it.
func TestMACDLinear(t *testing.T) {
	input := make([]float64, 40)
	for i := range input {
		input[i] = float64(i + 1)
	}
	got := apply(input, MustNewMACD(3, 6, 3).Process)
	if want := len(input) - (6 + 3 - 2); len(got) != want {
		t.Fatalf("length mismatch: got %d want %d", len(got), want)
	}
	for i, v := range got {
		if math.Abs(v.MACD-1.5) > 1e-9 || math.Abs(v.Signal-1.5) > 1e-9 || math.Abs(v.Hist) > 1e-9 {
			t.Fatalf("value mismatch at %d: got %+v want {MACD:1.5 Signal:1.5 Hist:0}", i, v)
		}
	}
}

func TestMACD(t *testing.T) {
	input := randomWalk(1, 300)
	for _, tc := range []struct{ fast, slow, signal int }{
		{12, 26, 9},
		{3, 6, 3},
		{5, 10, 1},
		{2, 3, 4},
	} {
		got := apply(input, MustNewMACD(tc.fast, tc.slow, tc.signal).Process)
		macd, signal, hist := talibMACD(input, tc.fast, tc.slow, tc.signal)
		gotMACD := make([]float64, len(got))
		gotSignal := make([]float64, len(got))
		gotHist := make([]float64, len(got))
		for i, v := range got {
			gotMACD[i], gotSignal[i], gotHist[i] = v.MACD, v.Signal, v.Hist
		}
		assertFloatSliceApprox(t, gotMACD, macd, 1e-9)
		assertFloatSliceApprox(t, gotSignal, signal, 1e-9)
		assertFloatSliceApprox(t, gotHist, hist, 1e-9)
	}
}

// talibMACD ports TA-Lib C's TA_INT_MACD for a whole-array call and returns
// outputs for inputs lookback..len(in)-1. go-talib's Macd is not used as the
// reference because it seeds the signal EMA with leading zeros.
func talibMACD(in []float64, fast, slow, signal int) (macd, sig, hist []float64) {
	lookbackSignal := signal - 1
	start := slow - 1
	fastEMA := talibEMA(in, start, fast)
	slowEMA := talibEMA(in, start, slow)
	line := make([]float64, len(slowEMA))
	for i := range line {
		line[i] = fastEMA[i] - slowEMA[i]
	}
	sig = talibEMA(line, lookbackSignal, signal)
	macd = line[lookbackSignal:]
	hist = make([]float64, len(macd))
	for i := range hist {
		hist[i] = macd[i] - sig[i]
	}
	return macd, sig, hist
}

// talibEMA ports TA-Lib C's TA_INT_EMA: outputs for in[startIdx:], seeded with
// the mean of the period values ending at startIdx.
func talibEMA(in []float64, startIdx, period int) []float64 {
	k := 2 / (float64(period) + 1)
	sum := 0.0
	for _, v := range in[startIdx-period+1 : startIdx+1] {
		sum += v
	}
	prev := sum / float64(period)
	out := []float64{prev}
	for _, v := range in[startIdx+1:] {
		prev = (v-prev)*k + prev
		out = append(out, prev)
	}
	return out
}
