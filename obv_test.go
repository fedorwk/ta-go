package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestOBV(t *testing.T) {
	input := randomCandles(1, 300)
	// Repeat some closes so the unchanged-close branch is exercised.
	for i := 10; i < len(input); i += 10 {
		input[i].Close = input[i-1].Close
	}
	got := apply(input, MustNewOBV().Process)

	closes := make([]float64, len(input))
	volumes := make([]float64, len(input))
	for i, v := range input {
		closes[i] = v.Close
		volumes[i] = v.Volume
	}
	assertAlignedApprox(t, got, talib.Obv(closes, volumes), 0, 1e-9)
}
