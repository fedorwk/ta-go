package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestATR(t *testing.T) {
	input := []OHLC{
		{Open: 10, High: 12, Low: 9, Close: 11},
		{Open: 11, High: 13, Low: 10, Close: 12},
		{Open: 12, High: 14, Low: 11, Close: 13},
		{Open: 13, High: 15, Low: 12, Close: 14},
		{Open: 14, High: 16, Low: 13, Close: 15},
	}
	got := apply(input, MustNewATR(3).Process)
	high, low, close := ohlcToHighLowClose(input)
	want := tail(talib.Atr(high, low, close, 3), len(got))
	assertFloatSliceApprox(t, got, want, 1e-9)
}
