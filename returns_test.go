package ta

import (
	"errors"
	"fmt"
	"math"
	"testing"

	talib "github.com/markcheno/go-talib"
)

// bruteLogReturn spells out the LogReturn rule: ln(x/prev), or 0 unless both
// are positive.
func bruteLogReturn(x, prev float64) float64 {
	if x > 0 && prev > 0 {
		return math.Log(x / prev)
	}
	return 0
}

// bruteRealizedVol returns, for each input from index period on, the root of
// the sum of squared one-step log returns ending there, recomputed from
// scratch; earlier entries are 0.
func bruteRealizedVol(input []float64, period int) []float64 {
	out := make([]float64, len(input))
	for i := period; i < len(input); i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			r := bruteLogReturn(input[j], input[j-1])
			sum += r * r
		}
		out[i] = math.Sqrt(sum)
	}
	return out
}

// assertRelApprox checks got against want[lookback:] within a relative
// tolerance.
func assertRelApprox(t *testing.T, got, want []float64, lookback int, rel float64) {
	t.Helper()
	if len(got) != len(want)-lookback {
		t.Fatalf("first output at input index %d, want %d", len(want)-len(got), lookback)
	}
	for i, g := range got {
		w := want[i+lookback]
		if math.Abs(g-w) > rel*math.Abs(w) {
			t.Fatalf("output %d: got %.17g, want %.17g (rel diff %g > %g)", i, g, w, math.Abs(g-w)/math.Abs(w), rel)
		}
	}
}

func TestLogReturn(t *testing.T) {
	input := randomWalk(1, 300)
	for i, v := range input {
		if v <= 0 {
			t.Fatalf("input %d is %v: the reference needs positive values", i, v)
		}
	}
	for _, period := range []int{1, 2, 10, 50} {
		rocr := talib.Rocr(input, period)
		want := make([]float64, len(input))
		for i := period; i < len(input); i++ {
			want[i] = math.Log(rocr[i])
		}
		got := apply(input, MustNewLogReturn(period).Process)
		assertAlignedApprox(t, got, want, period, 1e-12)
	}
}

func TestLogReturnEdges(t *testing.T) {
	cases := []struct {
		name   string
		period int
		input  []float64
		want   []float64
	}{
		{"period 1 non-positive", 1, []float64{2, 0, 3, -1, 4, 4, 8}, []float64{0, 0, 0, 0, 0, math.Ln2}},
		{"period 2 non-positive", 2, []float64{2, 0, 4, -1, 8, 8}, []float64{math.Ln2, 0, math.Ln2, 0}},
		{"flat", 3, []float64{5, 5, 5, 5, 5}, []float64{0, 0}},
		{"zeros", 1, []float64{0, 0, 0}, []float64{0, 0}},
		{"NaN", 1, []float64{1, math.NaN(), 2}, []float64{0, 0}},
		{"ratio overflows", 1, []float64{1e-300, 1e300}, []float64{600 * math.Ln10}},
		{"ratio underflows", 1, []float64{1e300, 1e-300}, []float64{-600 * math.Ln10}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := apply(c.input, MustNewLogReturn(c.period).Process)
			assertFloatSliceApprox(t, got, c.want, 1e-12*math.Max(1, math.Abs(c.want[len(c.want)-1])))
		})
	}

	// A subnormal ratio would lose precision; a huge one would be Inf.
	for _, in := range [][2]float64{{5e-324, 1e308}, {1e308, 5e-324}, {math.MaxFloat64, math.SmallestNonzeroFloat64}} {
		got := apply(in[:], MustNewLogReturn(1).Process)
		want := math.Log(in[1]) - math.Log(in[0])
		if len(got) != 1 || math.IsInf(got[0], 0) || math.IsNaN(got[0]) || math.Abs(got[0]-want) > 1e-12*math.Abs(want) {
			t.Errorf("LogReturn(%g -> %g) = %v, want %v", in[0], in[1], got, want)
		}
	}
}

func TestRealizedVol(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewRealizedVol(period).Process)
				assertAlignedApprox(t, got, bruteRealizedVol(input, period), period, 1e-10)
			})
		}
	}
}

// TestRealizedVolSpikeThenFlat checks that a flat window returns exactly 0
// even right after large returns have left it, before the ring wraps.
func TestRealizedVolSpikeThenFlat(t *testing.T) {
	for _, period := range []int{1, 2, 3, 7, 20} {
		input := randomWalk(4, 3*period)
		input[period] *= 1.5 // a large return in and out
		last := input[len(input)-1]
		for range 3 * period {
			input = append(input, last)
		}
		got := apply(input, MustNewRealizedVol(period).Process)
		// The last nonzero return is at input index 3*period-1; it leaves the
		// window period inputs later.
		for i := 4*period - 1; i < len(input); i++ {
			if g := got[i-period]; g != 0 {
				t.Fatalf("period %d input %d: got %g, want exactly 0", period, i, g)
			}
		}
	}
}

// TestRealizedVolCollapse feeds large returns, then tiny steps near 1e5,
// where the residue of the large sum would dominate the small one.
func TestRealizedVolCollapse(t *testing.T) {
	var input []float64
	for i := range 40 {
		input = append(input, 1e5*(1+0.2*float64(i%2)))
	}
	for i := range 200 {
		input = append(input, 1e5+float64(i%7)*1e-3)
	}
	for _, period := range []int{2, 5, 20, 61} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			got := apply(input, MustNewRealizedVol(period).Process)
			assertRelApprox(t, got, bruteRealizedVol(input, period), period, 1e-9)
		})
	}
}

// TestRealizedVolLargeOffset checks a long stream near 1e5, where returns
// are around 1e-5, against a from-scratch sum for every window.
func TestRealizedVolLargeOffset(t *testing.T) {
	input := randomWalk(2, 100_000)
	for i := range input {
		input[i] += 1e5
	}
	sq := make([]float64, len(input))
	for i := 1; i < len(input); i++ {
		r := bruteLogReturn(input[i], input[i-1])
		sq[i] = r * r
	}
	for _, period := range []int{2, 14, 50, 390} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			want := make([]float64, len(input))
			for i := period; i < len(input); i++ {
				sum := 0.0
				for _, s := range sq[i-period+1 : i+1] {
					sum += s
				}
				want[i] = math.Sqrt(sum)
			}
			got := apply(input, MustNewRealizedVol(period).Process)
			assertRelApprox(t, got, want, period, 1e-9)
		})
	}
}

func TestReturnsInvalidPeriod(t *testing.T) {
	constructors := map[string]func(int) error{
		"LogReturn":   func(p int) error { _, err := NewLogReturn(p); return err },
		"RealizedVol": func(p int) error { _, err := NewRealizedVol(p); return err },
	}
	for name, newX := range constructors {
		for _, p := range []int{0, -1} {
			if err := newX(p); !errors.Is(err, ErrInvalidPeriod) {
				t.Errorf("New%s(%d): got %v, want ErrInvalidPeriod", name, p, err)
			}
		}
	}
	mustConstructors := map[string]func(){
		"LogReturn":   func() { MustNewLogReturn(0) },
		"RealizedVol": func() { MustNewRealizedVol(0) },
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

func TestReturnsNoAlloc(t *testing.T) {
	input := randomWalk(1, 300)
	processors := map[string]Processor[float64, float64]{
		"LogReturn":   MustNewLogReturn(14),
		"RealizedVol": MustNewRealizedVol(14),
	}
	for name, p := range processors {
		i := 0
		allocs := testing.AllocsPerRun(len(input)-1, func() {
			p.Process(input[i%len(input)])
			i++
		})
		if allocs != 0 {
			t.Errorf("%s.Process: %v allocs per call, want 0", name, allocs)
		}
	}
}
