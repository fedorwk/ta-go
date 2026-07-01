package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestBollinger(t *testing.T) {
	input := []float64{1, 2, 3, 4, 5}
	got := apply(input, MustNewBollinger(3, 2).Process)
	upper, middle, lower := talib.BBands(input, 3, 2, 2, talib.SMA)
	want := make([]Bands, len(got))
	upper = tail(upper, len(got))
	middle = tail(middle, len(got))
	lower = tail(lower, len(got))
	for i := range upper {
		want[i] = Bands{
			Upper:  upper[i],
			Middle: middle[i],
			Lower:  lower[i],
		}
	}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d want %d", len(got), len(want))
	}
	for i := range got {
		assertFloatSliceApprox(t, []float64{got[i].Upper, got[i].Middle, got[i].Lower}, []float64{want[i].Upper, want[i].Middle, want[i].Lower}, 1e-9)
	}
}
