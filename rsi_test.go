package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestRSI(t *testing.T) {
	input := []float64{44, 44.15, 43.9, 44.35, 44.75, 45.2, 45.4, 45.1, 45.5}
	got := apply(input, MustNewRSI(3).Process)
	want := tail(talib.Rsi(input, 3), len(got))
	assertFloatSliceApprox(t, got, want, 1e-9)
}
