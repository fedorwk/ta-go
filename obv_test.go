package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestOBV(t *testing.T) {
	input := []OHLCV{
		{Open: 10, High: 12, Low: 9, Close: 10, Volume: 100},
		{Open: 10, High: 13, Low: 10, Close: 11, Volume: 200},
		{Open: 11, High: 12, Low: 9, Close: 10, Volume: 300},
		{Open: 10, High: 11, Low: 9, Close: 10, Volume: 400},
		{Open: 10, High: 14, Low: 10, Close: 12, Volume: 500},
	}
	got := apply(input, MustNewOBV().Process)

	closes := make([]float64, len(input))
	volumes := make([]float64, len(input))
	for i, v := range input {
		closes[i] = v.Close
		volumes[i] = v.Volume
	}
	want := tail(talib.Obv(closes, volumes), len(got))
	assertFloatSliceApprox(t, got, want, 1e-9)
}
