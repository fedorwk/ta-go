package ta

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

func TestVWAP(t *testing.T) {
	input := randomCandles(1, 300)
	got := apply(input, MustNewVWAP().Process)
	want := vwapReference(input)
	assertAlignedApprox(t, got, want, 0, 1e-9)
}

func vwapReference(input []OHLCV) []float64 {
	result := make([]float64, len(input))
	cumulativePV := 0.0
	cumulativeVolume := 0.0

	for i, v := range input {
		typicalPrice := (v.High + v.Low + v.Close) / 3
		cumulativePV += typicalPrice * v.Volume
		cumulativeVolume += v.Volume
		if cumulativeVolume == 0 {
			result[i] = 0
			continue
		}
		result[i] = cumulativePV / cumulativeVolume
	}

	return result
}

// Reset starts a new session: VWAP then covers only the session's candles.
func TestVWAPSession(t *testing.T) {
	input := randomCandles(1, 300)
	vwap := MustNewVWAP()
	for start := 0; start < len(input); start += 100 {
		session := input[start : start+100]
		vwap.Reset()
		got := apply(session, vwap.Process)
		assertAlignedApprox(t, got, vwapReference(session), 0, 1e-9)
	}
}

// bruteVWAPPeriod returns, for each candle from index period-1 on, the VWAP
// of the period candles ending there, summed from scratch, or the candle's
// typical price when they traded nothing; earlier entries are 0.
func bruteVWAPPeriod(candles []OHLCV, period int) []float64 {
	out := make([]float64, len(candles))
	for i := period - 1; i < len(candles); i++ {
		pv, vol := 0.0, 0.0
		for _, c := range candles[i-period+1 : i+1] {
			pv += typicalPrice(c) * c.Volume
			vol += c.Volume
		}
		if vol <= 0 {
			out[i] = typicalPrice(candles[i])
			continue
		}
		out[i] = pv / vol
	}
	return out
}

func typicalPrice(c OHLCV) float64 {
	return (c.High + c.Low + c.Close) / 3
}

// shiftCandles adds offset to every price of the candles, in place.
func shiftCandles(candles []OHLCV, offset float64) []OHLCV {
	for i := range candles {
		candles[i].Open += offset
		candles[i].High += offset
		candles[i].Low += offset
		candles[i].Close += offset
	}
	return candles
}

func TestVWAPPeriod(t *testing.T) {
	input := randomCandles(1, 300)
	for _, period := range orderStatTestPeriods {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			got := apply(input, MustNewVWAPPeriod(period).Process)
			assertRelApprox(t, got, bruteVWAPPeriod(input, period), period-1, 1e-12)
			// The first window is the cumulative VWAP of its candles.
			if want := vwapReference(input)[period-1]; math.Abs(got[0]-want) > 1e-12*want {
				t.Errorf("first output %.17g, want cumulative VWAP %.17g", got[0], want)
			}
		})
	}
}

// TestVWAPPeriodOne checks that a period of 1 returns each candle's typical
// price.
func TestVWAPPeriodOne(t *testing.T) {
	input := randomCandles(2, 300)
	input[10].Volume = 0
	got := apply(input, MustNewVWAPPeriod(1).Process)
	for i, c := range input {
		if want := typicalPrice(c); math.Abs(got[i]-want) > 1e-12*want {
			t.Fatalf("candle %d: got %.17g, want %.17g", i, got[i], want)
		}
	}
}

func TestVWAPPeriodZeroVolume(t *testing.T) {
	for _, period := range []int{1, 2, 3, 7, 20} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			// A window that never traded returns each candle's typical price.
			flat := randomCandles(3, 5*period)
			for i := range flat {
				flat[i].Volume = 0
			}
			got := apply(flat, MustNewVWAPPeriod(period).Process)
			for i, g := range got {
				if want := typicalPrice(flat[i+period-1]); g != want {
					t.Fatalf("no volume, candle %d: got %.17g, want exactly %.17g", i+period-1, g, want)
				}
			}

			// Heavy trading, then none: once the traded candles leave the
			// window, before and after it wraps, the output is exactly the
			// typical price.
			input := shiftCandles(randomCandles(4, 3*period), 1e5)
			for i := range input {
				input[i].Volume *= 1e6
			}
			quiet := shiftCandles(randomCandles(5, 3*period), 1e5)
			for i := range quiet {
				quiet[i].Volume = 0
			}
			input = append(input, quiet...)
			got = apply(input, MustNewVWAPPeriod(period).Process)
			assertRelApprox(t, got, bruteVWAPPeriod(input, period), period-1, 1e-12)
			for i := 4*period - 1; i < len(input); i++ {
				if g, want := got[i-period+1], typicalPrice(input[i]); g != want {
					t.Fatalf("after trading stops, candle %d: got %.17g, want exactly %.17g", i, g, want)
				}
			}

			// Volume that comes and goes.
			mixed := randomCandles(6, 300)
			for i := range mixed {
				if i%5 < 3 {
					mixed[i].Volume = 0
				}
			}
			got = apply(mixed, MustNewVWAPPeriod(period).Process)
			assertRelApprox(t, got, bruteVWAPPeriod(mixed, period), period-1, 1e-12)
		})
	}
}

// TestVWAPPeriodCollapse feeds heavy volume, then light volume, near 1e5,
// where the residue of the large sums would dominate the small ones.
func TestVWAPPeriodCollapse(t *testing.T) {
	input := shiftCandles(randomCandles(7, 240), 1e5)
	for i := range 40 {
		input[i].Volume *= 1e9
	}
	for _, period := range []int{2, 5, 20, 61} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			got := apply(input, MustNewVWAPPeriod(period).Process)
			assertRelApprox(t, got, bruteVWAPPeriod(input, period), period-1, 1e-12)
		})
	}
}

// TestVWAPPeriodLargeOffset checks a long stream near 1e5 against a
// from-scratch sum for every window.
func TestVWAPPeriodLargeOffset(t *testing.T) {
	input := shiftCandles(randomCandles(8, 100_000), 1e5)
	for _, period := range []int{2, 14, 390} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			got := apply(input, MustNewVWAPPeriod(period).Process)
			assertRelApprox(t, got, bruteVWAPPeriod(input, period), period-1, 1e-12)
		})
	}
}

func TestVWAPPeriodInvalidPeriod(t *testing.T) {
	for _, p := range []int{0, -1} {
		if _, err := NewVWAPPeriod(p); !errors.Is(err, ErrInvalidPeriod) {
			t.Errorf("NewVWAPPeriod(%d): got %v, want ErrInvalidPeriod", p, err)
		}
	}
	defer func() {
		if recover() == nil {
			t.Error("MustNewVWAPPeriod(0) did not panic")
		}
	}()
	MustNewVWAPPeriod(0)
}

func TestVWAPPeriodNoAlloc(t *testing.T) {
	assertNoAlloc(t, "VWAPPeriod", Processor[OHLCV, float64](MustNewVWAPPeriod(14)), randomCandles(1, 300))
}
