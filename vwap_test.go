package ta

import "testing"

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
