package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestStoch(t *testing.T) {
	input := []OHLC{
		{Open: 10, High: 12, Low: 9, Close: 11},
		{Open: 11, High: 13, Low: 10, Close: 12},
		{Open: 12, High: 14, Low: 11, Close: 13},
		{Open: 11, High: 15, Low: 10, Close: 14},
		{Open: 13, High: 16, Low: 12, Close: 15},
		{Open: 14, High: 17, Low: 13, Close: 16},
	}
	got := apply(input, MustNewStoch(3, 3).Process)
	high, low, close := ohlcToHighLowClose(input)
	k, d := talib.Stoch(high, low, close, 3, 1, talib.SMA, 3, talib.SMA)
	want := make([]StochValue, len(got))
	k = tail(k, len(got))
	d = tail(d, len(got))
	for i := range k {
		want[i] = StochValue{K: k[i], D: d[i]}
	}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d want %d", len(got), len(want))
	}
	for i := range got {
		assertFloatSliceApprox(t, []float64{got[i].K, got[i].D}, []float64{want[i].K, want[i].D}, 1e-9)
	}
}
