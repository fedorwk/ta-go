package ta

import (
	"math"
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestMACD(t *testing.T) {
	input := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	got := apply(input, MustNewMACD(3, 6, 3).Process)
	macd, signal, hist := talib.Macd(input, 3, 6, 3)
	want := make([]MACDValue, len(got))
	macd = tail(macd, len(got))
	signal = tail(signal, len(got))
	hist = tail(hist, len(got))
	for i := range macd {
		want[i] = MACDValue{
			MACD:   macd[i],
			Signal: signal[i],
			Hist:   hist[i],
		}
	}
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got %d want %d", len(got), len(want))
	}
	for i := range got {
		if math.Abs(got[i].MACD-want[i].MACD) > 1e-9 {
			t.Fatalf("macd mismatch at %d: got %.8f want %.8f", i, got[i].MACD, want[i].MACD)
		}
		if math.Abs(got[i].Signal-want[i].Signal) > 1e-9 {
			t.Fatalf("signal mismatch at %d: got %.8f want %.8f", i, got[i].Signal, want[i].Signal)
		}
		if math.Abs(got[i].Hist-want[i].Hist) > 1e-9 {
			t.Fatalf("hist mismatch at %d: got %.8f want %.8f", i, got[i].Hist, want[i].Hist)
		}
	}
}
