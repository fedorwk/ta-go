package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestSMA(t *testing.T) {
	input := []float64{1, 2, 3, 4, 5}
	got := apply(input, MustNewSMA(3).Process)
	want := tail(talib.Sma(input, 3), len(got))
	assertFloatSliceApprox(t, got, want, 1e-9)
}
