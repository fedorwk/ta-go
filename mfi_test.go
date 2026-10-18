package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestMFI(t *testing.T) {
	input := randomCandles(1, 300)
	high := make([]float64, len(input))
	low := make([]float64, len(input))
	close := make([]float64, len(input))
	volume := make([]float64, len(input))
	for i, v := range input {
		high[i], low[i], close[i], volume[i] = v.High, v.Low, v.Close, v.Volume
	}
	for _, period := range []int{1, 2, 3, 14, 50} {
		got := apply(input, MustNewMFI(period).Process)
		assertAlignedApprox(t, got, talib.Mfi(high, low, close, volume, period), period, 1e-9)
	}
}
