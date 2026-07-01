package ta

import "testing"

func TestVWAP(t *testing.T) {
	input := []OHLCV{
		{Open: 10, High: 12, Low: 9, Close: 11, Volume: 100},
		{Open: 11, High: 13, Low: 10, Close: 12, Volume: 200},
		{Open: 12, High: 14, Low: 11, Close: 13, Volume: 300},
		{Open: 13, High: 16, Low: 12, Close: 14, Volume: 400},
		{Open: 14, High: 17, Low: 13, Close: 15, Volume: 500},
	}
	got := apply(input, MustNewVWAP().Process)
	want := vwapReference(input)
	assertFloatSliceApprox(t, got, want, 1e-9)
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
