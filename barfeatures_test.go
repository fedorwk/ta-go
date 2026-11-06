package ta

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"

	talib "github.com/markcheno/go-talib"
)

func bruteCLV(v OHLC) float64 {
	if v.High == v.Low {
		return 0.5
	}
	return math.Min(math.Max((v.Close-v.Low)/(v.High-v.Low), 0), 1)
}

// bruteEfficiencyRatio returns, for each input from index period on, the net
// change over the last period steps divided by their path length; earlier
// entries are 0.
func bruteEfficiencyRatio(input []float64, period int) []float64 {
	out := make([]float64, len(input))
	for i := period; i < len(input); i++ {
		path := 0.0
		for j := i - period + 1; j <= i; j++ {
			path += math.Abs(input[j] - input[j-1])
		}
		if path > 0 {
			out[i] = math.Min(math.Abs(input[i]-input[i-period])/path, 1)
		}
	}
	return out
}

// bruteRatioToPrior returns, for each index from first on, x[i] divided by
// the median of the period values before it, or 0 when that median is ≤ 0;
// earlier entries are 0.
func bruteRatioToPrior(x []float64, first, period int) []float64 {
	out := make([]float64, len(x))
	for i := first; i < len(x); i++ {
		if base := bruteMedian(x[i-period : i]); base > 0 {
			out[i] = x[i] / base
		}
	}
	return out
}

// bruteTrueRanges returns the true range of each candle from the 2nd on;
// entry 0 is 0.
func bruteTrueRanges(candles []OHLC) []float64 {
	out := make([]float64, len(candles))
	for i := 1; i < len(candles); i++ {
		c, prev := candles[i], candles[i-1].Close
		out[i] = math.Max(c.High-c.Low, math.Max(math.Abs(c.High-prev), math.Abs(c.Low-prev)))
	}
	return out
}

// flatCandles turns each value into a candle with every price equal to it,
// so a tie-heavy series gives tie-heavy and often zero true ranges.
func flatCandles(values []float64) []OHLC {
	out := make([]OHLC, len(values))
	for i, v := range values {
		out[i] = OHLC{Open: v, High: v, Low: v, Close: v}
	}
	return out
}

func TestCLV(t *testing.T) {
	candles := randomOHLC(1, 300)
	want := make([]float64, len(candles))
	for i, c := range candles {
		want[i] = bruteCLV(c)
	}
	got := apply(candles, MustNewCLV().Process)
	assertAlignedApprox(t, got, want, 0, 0)
}

// TestCLVAd checks that (2·CLV − 1)·V is one step of TA-Lib's AD line.
func TestCLVAd(t *testing.T) {
	candles := randomCandles(2, 300)
	high, low, close, vol := make([]float64, len(candles)), make([]float64, len(candles)), make([]float64, len(candles)), make([]float64, len(candles))
	for i, c := range candles {
		if c.High == c.Low {
			t.Fatalf("candle %d has H = L: the reference needs a range", i)
		}
		high[i], low[i], close[i], vol[i] = c.High, c.Low, c.Close, c.Volume
	}
	ad := talib.Ad(high, low, close, vol)
	clv := MustNewCLV()
	for i, c := range candles {
		got, ok := clv.Process(OHLC{Open: c.Open, High: c.High, Low: c.Low, Close: c.Close})
		if !ok {
			t.Fatalf("candle %d not ready", i)
		}
		step := ad[i]
		if i > 0 {
			step -= ad[i-1]
		}
		if d := (2*got-1)*c.Volume - step; math.Abs(d) > 1e-8 {
			t.Fatalf("candle %d: (2·CLV−1)·V = %g, AD step %g", i, (2*got-1)*c.Volume, step)
		}
	}
}

func TestCLVEdges(t *testing.T) {
	cases := []struct {
		name string
		in   OHLC
		want float64
	}{
		{"H = L", OHLC{5, 5, 5, 5}, 0.5},
		{"zeros", OHLC{}, 0.5},
		{"close at high", OHLC{1, 3, 1, 3}, 1},
		{"close at low", OHLC{2, 3, 1, 1}, 0},
		{"mid", OHLC{2, 4, 2, 3}, 0.5},
		{"close above high", OHLC{2, 3, 1, 4}, 1},
		{"close below low", OHLC{2, 3, 1, 0}, 0},
	}
	for _, c := range cases {
		if got, ok := MustNewCLV().Process(c.in); !ok || got != c.want {
			t.Errorf("%s: got %v, %v; want %v", c.name, got, ok, c.want)
		}
	}
}

func TestEfficiencyRatio(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewEfficiencyRatio(period).Process)
				assertAlignedApprox(t, got, bruteEfficiencyRatio(input, period), period, 1e-10)
			})
		}
	}
}

func TestEfficiencyRatioEdges(t *testing.T) {
	// Integer steps keep the arithmetic exact, so a one-way series scores
	// exactly 1 and a flat one exactly 0.
	up := []float64{0}
	for i := range 100 {
		up = append(up, up[len(up)-1]+float64(1+i%3))
	}
	for _, period := range []int{1, 2, 7, 20} {
		for i, g := range apply(up, MustNewEfficiencyRatio(period).Process) {
			if g != 1 {
				t.Fatalf("one-way period %d output %d: got %g, want 1", period, i, g)
			}
		}
		flat := make([]float64, 50)
		for i, g := range apply(flat, MustNewEfficiencyRatio(period).Process) {
			if g != 0 {
				t.Fatalf("flat period %d output %d: got %g, want 0", period, i, g)
			}
		}
	}

	// Back and forth over an even number of steps ends where it started.
	chop := []float64{1, 2, 1, 2, 1, 2, 1}
	if got := apply(chop, MustNewEfficiencyRatio(2).Process); !allEqual(got, 0) {
		t.Errorf("chop period 2: got %v, want all 0", got)
	}
	// Period 1 is 1 for any change and 0 for none.
	if got, want := apply([]float64{1, 3, 3, 2}, MustNewEfficiencyRatio(1).Process), []float64{1, 0, 1}; !slices.Equal(got, want) {
		t.Errorf("period 1: got %v, want %v", got, want)
	}
}

// TestEfficiencyRatioSpikeThenFlat checks that a flat window returns exactly
// 0 even right after large moves have left it, before the ring wraps.
func TestEfficiencyRatioSpikeThenFlat(t *testing.T) {
	for _, period := range []int{1, 2, 3, 7, 20} {
		input := randomWalk(4, 3*period)
		input[period] *= 1.5
		last := input[len(input)-1]
		for range 3 * period {
			input = append(input, last)
		}
		got := apply(input, MustNewEfficiencyRatio(period).Process)
		for i := 4*period - 1; i < len(input); i++ {
			if g := got[i-period]; g != 0 {
				t.Fatalf("period %d input %d: got %g, want exactly 0", period, i, g)
			}
		}
	}
}

// TestEfficiencyRatioCollapse feeds large moves, then tiny steps near 1e5,
// where the residue of the large path length would dominate the small one.
func TestEfficiencyRatioCollapse(t *testing.T) {
	// Random large moves leave rounding residue in the running path length;
	// moves that are exact binary values would not.
	r := rand.New(rand.NewSource(5))
	var input []float64
	for range 40 {
		input = append(input, 1e5*(1+0.2*r.Float64()))
	}
	for i := range 200 {
		input = append(input, 1e5+float64(i%7+i/5)*1e-3)
	}
	for _, period := range []int{2, 5, 20, 61} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			got := apply(input, MustNewEfficiencyRatio(period).Process)
			assertRelApprox(t, got, bruteEfficiencyRatio(input, period), period, 1e-9)
		})
	}
}

func TestRangeExpansion(t *testing.T) {
	inputs := map[string][]OHLC{"walk": randomOHLC(1, 300)}
	for name, values := range orderStatInputs() {
		inputs[name] = flatCandles(values)
	}
	for name, candles := range inputs {
		trs := bruteTrueRanges(candles)
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(candles, MustNewRangeExpansion(period).Process)
				// The first output is on candle period+2, input index period+1.
				assertAlignedApprox(t, got, bruteRatioToPrior(trs, period+1, period), period+1, 1e-12)
			})
		}
	}
}

func TestRangeExpansionEdges(t *testing.T) {
	flat := flatCandles(make([]float64, 30))
	if got := apply(flat, MustNewRangeExpansion(5).Process); len(got) != 30-6 || !allEqual(got, 0) {
		t.Errorf("flat: got %v, want %d zeros", got, 30-6)
	}
	// Ranges of 2 then one of 6 expand threefold.
	candles := []OHLC{{10, 11, 9, 10}, {10, 11, 9, 10}, {10, 11, 9, 10}, {10, 11, 9, 10}, {10, 13, 7, 10}}
	if got := apply(candles, MustNewRangeExpansion(3).Process); !slices.Equal(got, []float64{3}) {
		t.Errorf("expansion: got %v, want [3]", got)
	}
}

func TestRelVolume(t *testing.T) {
	inputs := map[string][]OHLCV{"walk": randomCandles(1, 300)}
	for name, values := range orderStatInputs() {
		candles := make([]OHLCV, len(values))
		for i, v := range values {
			candles[i] = OHLCV{Open: 1, High: 1, Low: 1, Close: 1, Volume: math.Abs(v)}
		}
		inputs[name] = candles
	}
	for name, candles := range inputs {
		vols := make([]float64, len(candles))
		for i, c := range candles {
			vols[i] = c.Volume
		}
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(candles, MustNewRelVolume(period).Process)
				assertAlignedApprox(t, got, bruteRatioToPrior(vols, period, period), period, 1e-12)
			})
		}
	}
}

func TestRelVolumeEdges(t *testing.T) {
	withVolumes := func(vols ...float64) []OHLCV {
		out := make([]OHLCV, len(vols))
		for i, v := range vols {
			out[i] = OHLCV{Volume: v}
		}
		return out
	}
	cases := []struct {
		name   string
		period int
		vols   []float64
		want   []float64
	}{
		{"spike in flat window", 3, []float64{100, 100, 100, 500, 100}, []float64{5, 1}},
		{"zero baseline", 3, []float64{0, 0, 7, 9, 4}, []float64{0, 4.0 / 7}},
		{"zero volume", 2, []float64{4, 4, 0}, []float64{0}},
		{"period 1", 1, []float64{2, 4, 1}, []float64{2, 0.25}},
	}
	for _, c := range cases {
		if got := apply(withVolumes(c.vols...), MustNewRelVolume(c.period).Process); !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}

	// A subnormal baseline would make the ratio infinite.
	got := apply(withVolumes(5e-324, 5e-324, 1e308), MustNewRelVolume(2).Process)
	if len(got) != 1 || got[0] != math.MaxFloat64 {
		t.Errorf("tiny baseline: got %v, want [MaxFloat64]", got)
	}
}

func TestBarFeaturesInvalidPeriod(t *testing.T) {
	constructors := map[string]func(int) error{
		"EfficiencyRatio": func(p int) error { _, err := NewEfficiencyRatio(p); return err },
		"RangeExpansion":  func(p int) error { _, err := NewRangeExpansion(p); return err },
		"RelVolume":       func(p int) error { _, err := NewRelVolume(p); return err },
	}
	for name, newX := range constructors {
		for _, p := range []int{0, -1} {
			if err := newX(p); !errors.Is(err, ErrInvalidPeriod) {
				t.Errorf("New%s(%d): got %v, want ErrInvalidPeriod", name, p, err)
			}
		}
	}
	mustConstructors := map[string]func(){
		"EfficiencyRatio": func() { MustNewEfficiencyRatio(0) },
		"RangeExpansion":  func() { MustNewRangeExpansion(0) },
		"RelVolume":       func() { MustNewRelVolume(0) },
	}
	for name, mustNewX := range mustConstructors {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("MustNew%s(0) did not panic", name)
				}
			}()
			mustNewX()
		}()
	}
}

// assertNoAlloc checks that p.Process does not allocate once constructed.
func assertNoAlloc[I, O any](t *testing.T, name string, p Processor[I, O], input []I) {
	t.Helper()
	i := 0
	allocs := testing.AllocsPerRun(len(input)-1, func() {
		p.Process(input[i%len(input)])
		i++
	})
	if allocs != 0 {
		t.Errorf("%s.Process: %v allocs per call, want 0", name, allocs)
	}
}

func TestBarFeaturesNoAlloc(t *testing.T) {
	assertNoAlloc(t, "CLV", Processor[OHLC, float64](MustNewCLV()), randomOHLC(1, 300))
	assertNoAlloc(t, "EfficiencyRatio", Processor[float64, float64](MustNewEfficiencyRatio(14)), randomWalk(1, 300))
	assertNoAlloc(t, "RangeExpansion", Processor[OHLC, float64](MustNewRangeExpansion(14)), randomOHLC(1, 300))
	assertNoAlloc(t, "RelVolume", Processor[OHLCV, float64](MustNewRelVolume(14)), randomCandles(1, 300))
}

func allEqual(s []float64, v float64) bool {
	for _, x := range s {
		if x != v {
			return false
		}
	}
	return true
}
