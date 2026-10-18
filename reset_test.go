package ta

import "testing"

type resettable[I, O any] interface {
	Processor[I, O]
	Reset()
}

// testReset checks that after any amount of input, Reset makes the
// processor's output identical to that of a fresh one.
func testReset[I any, O comparable](newP func() resettable[I, O], input []I) func(*testing.T) {
	return func(t *testing.T) {
		want := Apply(newP(), input)
		if len(want) == 0 {
			t.Fatal("no output to compare")
		}
		for _, n := range []int{0, 1, 7, len(input) / 2, len(input)} {
			p := newP()
			Apply(p, input[:n])
			p.Reset()
			got := Apply(p, input)
			if len(got) != len(want) {
				t.Fatalf("after %d inputs and Reset: got %d outputs, want %d", n, len(got), len(want))
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatalf("after %d inputs and Reset: output %d is %+v, want %+v", n, i, got[i], want[i])
				}
			}
		}
	}
}

func TestReset(t *testing.T) {
	values := randomWalk(1, 300)
	ohlc := randomOHLC(1, 300)
	ohlcv := randomCandles(1, 300)

	t.Run("SMA", testReset(func() resettable[float64, float64] { return MustNewSMA(14) }, values))
	t.Run("EMA", testReset(func() resettable[float64, float64] { return MustNewEMA(14) }, values))
	t.Run("WMA", testReset(func() resettable[float64, float64] { return MustNewWMA(14) }, values))
	t.Run("StdDev", testReset(func() resettable[float64, float64] { return MustNewStdDev(14) }, values))
	t.Run("RSI", testReset(func() resettable[float64, float64] { return MustNewRSI(14) }, values))
	t.Run("ROC", testReset(func() resettable[float64, float64] { return MustNewROC(10) }, values))
	t.Run("Mom", testReset(func() resettable[float64, float64] { return MustNewMom(10) }, values))
	t.Run("MACD", testReset(func() resettable[float64, MACDValue] { return MustNewMACD(12, 26, 9) }, values))
	t.Run("StochRSI", testReset(func() resettable[float64, StochValue] { return MustNewStochRSI(14, 14, 3) }, values))
	for _, mt := range allMATypes {
		t.Run("MA/"+mt.String(), testReset(func() resettable[float64, float64] { return MustNewMA(14, mt) }, values))
		t.Run("Bollinger/"+mt.String(), testReset(func() resettable[float64, Bands] { return MustNewBollingerMA(20, 2, mt) }, values))
		t.Run("StochFull/"+mt.String(), testReset(func() resettable[OHLC, StochValue] { return MustNewStochFull(14, 3, 3, mt) }, ohlc))
	}

	t.Run("Stoch", testReset(func() resettable[OHLC, StochValue] { return MustNewStoch(14, 3) }, ohlc))
	t.Run("ATR", testReset(func() resettable[OHLC, float64] { return MustNewATR(14) }, ohlc))
	t.Run("ADX", testReset(func() resettable[OHLC, ADXValue] { return MustNewADX(14) }, ohlc))
	t.Run("WillR", testReset(func() resettable[OHLC, float64] { return MustNewWillR(14) }, ohlc))
	t.Run("CCI", testReset(func() resettable[OHLC, float64] { return MustNewCCI(20) }, ohlc))
	t.Run("Donchian", testReset(func() resettable[OHLC, Bands] { return MustNewDonchian(20) }, ohlc))

	t.Run("OBV", testReset(func() resettable[OHLCV, float64] { return MustNewOBV() }, ohlcv))
	t.Run("VWAP", testReset(func() resettable[OHLCV, float64] { return MustNewVWAP() }, ohlcv))
	t.Run("MFI", testReset(func() resettable[OHLCV, float64] { return MustNewMFI(14) }, ohlcv))

	t.Run("Chain", testReset(func() resettable[float64, float64] {
		return Chain(MustNewRSI(14), MustNewEMA(5)).(resettable[float64, float64])
	}, values))
	t.Run("ChainSource", testReset(func() resettable[OHLC, float64] {
		return Chain(HLC3, MustNewSMA(20)).(resettable[OHLC, float64])
	}, ohlc))
}
