package ta

import "testing"

func TestChain(t *testing.T) {
	input := randomWalk(1, 300)
	got := Apply(Chain(MustNewRSI(14), MustNewSMA(5)), input)
	want := Apply(MustNewSMA(5), Apply(MustNewRSI(14), input))
	assertFloatSliceApprox(t, got, want, 0)
}

func TestSource(t *testing.T) {
	v := OHLC{Open: 1, High: 4, Low: 0, Close: 2}
	for _, tc := range []struct {
		src  Source
		want float64
	}{
		{Open, 1}, {High, 4}, {Low, 0}, {Close, 2},
		{HL2, 2}, {HLC3, 2}, {OHLC4, 1.75},
	} {
		if got, ok := tc.src.Process(v); !ok || got != tc.want {
			t.Errorf("Source %d: got %v, %v; want %v, true", tc.src, got, ok, tc.want)
		}
	}
}

func TestChainSource(t *testing.T) {
	input := randomOHLC(1, 300)
	got := Apply(Chain(HLC3, MustNewSMA(20)), input)
	typical := make([]float64, len(input))
	for i, v := range input {
		typical[i] = (v.High + v.Low + v.Close) / 3
	}
	want := Apply(MustNewSMA(20), typical)
	assertFloatSliceApprox(t, got, want, 0)
}
