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

func makePairs(xs, ys []float64) []Pair {
	out := make([]Pair, len(xs))
	for i := range xs {
		out[i] = Pair{X: xs[i], Y: ys[i]}
	}
	return out
}

func splitPairs(ps []Pair) ([]float64, []float64) {
	xs := make([]float64, len(ps))
	ys := make([]float64, len(ps))
	for i, p := range ps {
		xs[i], ys[i] = p.X, p.Y
	}
	return xs, ys
}

// pairInputs returns two independent walks, tie-heavy small integers, mostly
// zeros, and ties against a walk.
func pairInputs() map[string][]Pair {
	stats := orderStatInputs()
	r := rand.New(rand.NewSource(5))
	shuffled := func(v []float64) []float64 {
		s := slices.Clone(v)
		r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
		return s
	}
	return map[string][]Pair{
		"walks":     makePairs(randomWalk(1, 300), randomWalk(2, 300)),
		"ties":      makePairs(stats["ties"], shuffled(stats["ties"])),
		"zeros":     makePairs(stats["zeros"], shuffled(stats["zeros"])),
		"ties/walk": makePairs(stats["ties"], randomWalk(2, 300)),
	}
}

// pairMoments holds the means and centred sums of a window of pairs.
type pairMoments struct {
	n, mx, my, cxx, cyy, cxy float64
}

// brutePairMoments computes the moments of w with a two-pass computation. A
// side whose values are all equal gets that value as its mean and 0 for its
// centred sums.
func brutePairMoments(w []Pair) pairMoments {
	xs, ys := splitPairs(w)
	mx, cxx, flatX := bruteCentre(xs)
	my, cyy, flatY := bruteCentre(ys)
	m := pairMoments{n: float64(len(w)), mx: mx, my: my, cxx: cxx, cyy: cyy}
	if !flatX && !flatY {
		for i := range xs {
			m.cxy += (xs[i] - mx) * (ys[i] - my)
		}
	}
	return m
}

// bruteCentre returns the mean and Σ(v − mean)² of vs, and whether every
// value is equal.
func bruteCentre(vs []float64) (mean, css float64, flat bool) {
	if allEqual(vs, vs[0]) {
		return vs[0], 0, true
	}
	for _, v := range vs {
		mean += v
	}
	mean /= float64(len(vs))
	for _, v := range vs {
		css += (v - mean) * (v - mean)
	}
	return mean, css, false
}

func bruteRegression(m pairMoments) RegressionValue {
	if m.cxx == 0 {
		return RegressionValue{Intercept: m.my}
	}
	slope := m.cxy / m.cxx
	v := RegressionValue{Slope: slope, Intercept: m.my - slope*m.mx}
	if m.cyy > 0 {
		v.R2 = m.cxy * m.cxy / (m.cxx * m.cyy)
	}
	return v
}

func bruteCorrel(m pairMoments) float64 {
	if m.cxx == 0 || m.cyy == 0 {
		return 0
	}
	return m.cxy / math.Sqrt(m.cxx*m.cyy)
}

// bruteLinRegSlope fits w against positions 0..len(w)−1 with a two-pass
// computation; a flat window or a single value has slope 0.
func bruteLinRegSlope(w []float64) float64 {
	n := len(w)
	my, _, flat := bruteCentre(w)
	if flat {
		return 0
	}
	c := float64(n-1) / 2
	num, den := 0.0, 0.0
	for i, y := range w {
		num += (float64(i) - c) * (y - my)
		den += (float64(i) - c) * (float64(i) - c)
	}
	return num / den
}

// brutePairWindows returns the moments of each full window of input, one per
// window, in order.
func brutePairWindows(input []Pair, period int) []pairMoments {
	var out []pairMoments
	for i := period - 1; i < len(input); i++ {
		out = append(out, brutePairMoments(input[i-period+1:i+1]))
	}
	return out
}

// assertPairOutputs checks a Regression's and a Correl's outputs, one per full
// window, against the brute-force moments of each window. The tolerances are
// relative to the natural scale of each quantity: the slope to √(cyy/cxx),
// which is the slope a correlation of 1 would give, and the intercept to the
// sizes of ȳ and slope·x̄ and to the spreads behind them. Values that the
// rules fix exactly, such as a slope of 0 for a flat side, get a tolerance
// of 0.
func assertPairOutputs(t *testing.T, reg []RegressionValue, cor []float64, want []pairMoments, tol float64) {
	t.Helper()
	if len(reg) != len(want) || len(cor) != len(want) {
		t.Fatalf("got %d regression and %d correlation outputs, want %d", len(reg), len(cor), len(want))
	}
	for i, m := range want {
		w := bruteRegression(m)
		g := reg[i]
		scale := 0.0
		if m.cxx > 0 {
			scale = math.Abs(w.Slope) + math.Sqrt(m.cyy/m.cxx)
		}
		interceptTol := tol * (math.Abs(m.my) + math.Sqrt(m.cyy/m.n) + scale*(math.Abs(m.mx)+math.Sqrt(m.cxx/m.n)))
		if math.Abs(g.Slope-w.Slope) > tol*scale ||
			math.Abs(g.Intercept-w.Intercept) > interceptTol ||
			math.Abs(g.R2-w.R2) > tol*boolTol(m.cxx > 0 && m.cyy > 0) {
			t.Fatalf("Regression output %d: got %+v, want %+v (moments %+v)", i, g, w, m)
		}
		wc := bruteCorrel(m)
		if math.Abs(cor[i]-wc) > tol*boolTol(m.cxx > 0 && m.cyy > 0) {
			t.Fatalf("Correl output %d: got %.17g, want %.17g (moments %+v)", i, cor[i], wc, m)
		}
	}
}

// boolTol is 1 when a quantity is computed and 0 when the rules fix it.
func boolTol(computed bool) float64 {
	if computed {
		return 1
	}
	return 0
}

// assertLinRegSlopes checks got, one output per full window of input, against
// bruteLinRegSlope, within tol of the slope's natural scale
// √(Σ(y − ȳ)² / Σ(i − c)²).
func assertLinRegSlopes(t *testing.T, got, input []float64, period int, tol float64) {
	t.Helper()
	if len(got) != len(input)-period+1 {
		t.Fatalf("got %d outputs, want %d", len(got), len(input)-period+1)
	}
	n := float64(period)
	d := n * (n*n - 1) / 12
	for i, g := range got {
		w := input[i : i+period]
		want := bruteLinRegSlope(w)
		_, css, flat := bruteCentre(w)
		scale := 0.0
		if !flat {
			scale = math.Abs(want) + math.Sqrt(css/d)
		}
		if math.Abs(g-want) > tol*scale {
			t.Fatalf("output %d: got %.17g, want %.17g", i, g, want)
		}
	}
}

func TestLinRegSlopeTalib(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{2, 14, 50} {
		got := apply(input, MustNewLinRegSlope(period).Process)
		assertAlignedApprox(t, got, talib.LinearRegSlope(input, period), period-1, 1e-9)
	}
}

// TestCorrelTalib leaves out periods 2 and 3, where go-talib's naive sums
// are off by up to 1e-7 on these walks (two points correlate at exactly ±1,
// and go-talib gives -0.99999991). The brute-force test covers them.
func TestCorrelTalib(t *testing.T) {
	xs, ys := randomWalk(1, 300), randomWalk(2, 300)
	input := makePairs(xs, ys)
	for _, period := range []int{1, 14, 50} {
		got := apply(input, MustNewCorrel(period).Process)
		assertAlignedApprox(t, got, talib.Correl(xs, ys, period), period-1, 1e-9)
	}
}

func TestRegressionCorrel(t *testing.T) {
	for name, input := range pairInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				reg := apply(input, MustNewRegression(period).Process)
				cor := apply(input, MustNewCorrel(period).Process)
				assertPairOutputs(t, reg, cor, brutePairWindows(input, period), 1e-9)
			})
		}
	}
}

func TestLinRegSlope(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewLinRegSlope(period).Process)
				assertLinRegSlopes(t, got, input, period, 1e-9)
			})
		}
	}
}

// TestRegressionLargeOffset checks a long stream of correlated pairs, X near
// 1e5, against a two-pass computation of every window.
func TestRegressionLargeOffset(t *testing.T) {
	const n = 100_000
	a, b := randomWalk(7, n), randomWalk(8, n)
	xs, ys := make([]float64, n), make([]float64, n)
	for i := range a {
		xs[i] = 1e5 + a[i]
		ys[i] = b[i] + 0.5*a[i]
	}
	input := makePairs(xs, ys)
	for _, period := range []int{2, 14, 50, 390} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			reg := apply(input, MustNewRegression(period).Process)
			cor := apply(input, MustNewCorrel(period).Process)
			assertPairOutputs(t, reg, cor, brutePairWindows(input, period), 1e-9)
		})
	}
}

func TestLinRegSlopeLargeOffset(t *testing.T) {
	input := randomWalk(9, 100_000)
	for i := range input {
		input[i] += 1e5
	}
	for _, period := range []int{2, 14, 50, 390} {
		t.Run(fmt.Sprintf("period=%d", period), func(t *testing.T) {
			got := apply(input, MustNewLinRegSlope(period).Process)
			assertLinRegSlopes(t, got, input, period, 1e-9)
		})
	}
}

// TestRegressionCollapse feeds a wide X spread, then tiny steps near 1e5,
// where the residue of the large sums would dominate the small ones.
func TestRegressionCollapse(t *testing.T) {
	var xs []float64
	for i := range 40 {
		xs = append(xs, 1e5*(1+0.2*float64(i%2)))
	}
	for i := range 200 {
		xs = append(xs, 1e5+float64(i%7)*1e-3)
	}
	ys := randomWalk(6, len(xs))
	for side, input := range map[string][]Pair{"X": makePairs(xs, ys), "Y": makePairs(ys, xs)} {
		for _, period := range []int{2, 5, 20, 61} {
			t.Run(fmt.Sprintf("%s/period=%d", side, period), func(t *testing.T) {
				reg := apply(input, MustNewRegression(period).Process)
				cor := apply(input, MustNewCorrel(period).Process)
				assertPairOutputs(t, reg, cor, brutePairWindows(input, period), 1e-9)
			})
		}
	}
}

// TestRegressionTurnsFlat checks that once one side of the window has become
// constant, before the ring wraps, the results are exactly those the rules
// fix rather than rounding noise.
func TestRegressionTurnsFlat(t *testing.T) {
	for _, period := range []int{2, 3, 7, 20} {
		varied := randomWalk(10, 3*period)
		for i := range varied {
			varied[i] *= 1000
		}
		other := randomWalk(11, 6*period)
		flat := slices.Concat(varied, slices.Repeat([]float64{5}, 3*period))

		// X turns flat: slope 0, R2 0, Correl 0, intercept the mean of Y.
		input := makePairs(flat, other)
		reg := apply(input, MustNewRegression(period).Process)
		cor := apply(input, MustNewCorrel(period).Process)
		for i := 4*period - 1; i < len(input); i++ {
			g, c := reg[i-period+1], cor[i-period+1]
			my, _, _ := bruteCentre(other[i-period+1 : i+1])
			if g.Slope != 0 || g.R2 != 0 || c != 0 || math.Abs(g.Intercept-my) > 1e-12*math.Abs(my) {
				t.Fatalf("X flat, period %d input %d: got %+v and Correl %g, want {0 %g 0} and 0", period, i, g, c, my)
			}
		}

		// Y turns flat: slope 0, intercept exactly 5, R2 0, Correl 0.
		input = makePairs(other, flat)
		reg = apply(input, MustNewRegression(period).Process)
		cor = apply(input, MustNewCorrel(period).Process)
		for i := 4*period - 1; i < len(input); i++ {
			g, c := reg[i-period+1], cor[i-period+1]
			if g != (RegressionValue{Intercept: 5}) || c != 0 {
				t.Fatalf("Y flat, period %d input %d: got %+v and Correl %g, want {0 5 0} and 0", period, i, g, c)
			}
		}
	}
}

func TestRegressionEdges(t *testing.T) {
	walk := randomWalk(12, 200)
	line := func(m, c float64) []Pair {
		ps := make([]Pair, len(walk))
		for i, x := range walk {
			ps[i] = Pair{X: x, Y: m*x + c}
		}
		return ps
	}

	for _, period := range []int{2, 14, 50} {
		for _, c := range []struct {
			m, c, r float64
		}{{2, 3, 1}, {-0.5, 1, -1}} {
			input := line(c.m, c.c)
			for i, g := range apply(input, MustNewRegression(period).Process) {
				if math.Abs(g.Slope-c.m) > 1e-9 || math.Abs(g.Intercept-c.c) > 1e-7 || math.Abs(g.R2-1) > 1e-9 {
					t.Fatalf("line y=%gx+%g, period %d, output %d: got %+v", c.m, c.c, period, i, g)
				}
			}
			for i, g := range apply(input, MustNewCorrel(period).Process) {
				if math.Abs(g-c.r) > 1e-9 {
					t.Fatalf("line y=%gx+%g, period %d, output %d: Correl %g, want %g", c.m, c.c, period, i, g, c.r)
				}
			}
		}
	}

	// Constant X: the slope is undefined, so {0, ȳ, 0}.
	input := makePairs(slices.Repeat([]float64{7}, len(walk)), walk)
	for i, g := range apply(input, MustNewRegression(14).Process) {
		my, _, _ := bruteCentre(walk[i : i+14])
		if g.Slope != 0 || g.R2 != 0 || math.Abs(g.Intercept-my) > 1e-12*math.Abs(my) {
			t.Fatalf("constant X, output %d: got %+v, want {0 %g 0}", i, g, my)
		}
	}
	// Constant Y: {0, y, 0}.
	input = makePairs(walk, slices.Repeat([]float64{-3}, len(walk)))
	for i, g := range apply(input, MustNewRegression(14).Process) {
		if g != (RegressionValue{Intercept: -3}) {
			t.Fatalf("constant Y, output %d: got %+v, want {0 -3 0}", i, g)
		}
	}
	// Zeros.
	input = make([]Pair, 30)
	for i, g := range apply(input, MustNewRegression(5).Process) {
		if g != (RegressionValue{}) {
			t.Fatalf("zeros, output %d: got %+v, want {0 0 0}", i, g)
		}
	}
	// Period 1: {0, y, 0}.
	input = makePairs(walk, randomWalk(13, len(walk)))
	for i, g := range apply(input, MustNewRegression(1).Process) {
		if g != (RegressionValue{Intercept: input[i].Y}) {
			t.Fatalf("period 1, output %d: got %+v, want {0 %g 0}", i, g, input[i].Y)
		}
	}

	// Correl is 0 whenever a side is constant, for zeros, and for period 1.
	cases := map[string]struct {
		period int
		input  []Pair
	}{
		"constant X": {14, makePairs(slices.Repeat([]float64{7}, len(walk)), walk)},
		"constant Y": {14, makePairs(walk, slices.Repeat([]float64{-3}, len(walk)))},
		"zeros":      {5, make([]Pair, 30)},
		"period 1":   {1, makePairs(walk, randomWalk(13, len(walk)))},
	}
	for name, c := range cases {
		for i, g := range apply(c.input, MustNewCorrel(c.period).Process) {
			if g != 0 {
				t.Fatalf("Correl, %s, output %d: got %g, want 0", name, i, g)
			}
		}
	}
}

func TestLinRegSlopeEdges(t *testing.T) {
	// A straight line has its own slope.
	input := make([]float64, 100)
	for i := range input {
		input[i] = 3*float64(i) + 1
	}
	for _, period := range []int{2, 14, 50} {
		for i, g := range apply(input, MustNewLinRegSlope(period).Process) {
			if math.Abs(g-3) > 1e-12 {
				t.Fatalf("line, period %d, output %d: got %.17g, want 3", period, i, g)
			}
		}
	}

	// Period 1, a flat series and zeros return exactly 0.
	cases := map[string]struct {
		period int
		input  []float64
	}{
		"period 1": {1, randomWalk(14, 50)},
		"flat":     {14, slices.Repeat([]float64{1e5 + 0.1}, 50)},
		"zeros":    {5, make([]float64, 30)},
	}
	for name, c := range cases {
		for i, g := range apply(c.input, MustNewLinRegSlope(c.period).Process) {
			if g != 0 {
				t.Fatalf("%s, output %d: got %g, want 0", name, i, g)
			}
		}
	}

	// A series that turns flat, before the ring wraps, returns exactly 0
	// once the window is flat.
	for _, period := range []int{2, 3, 7, 20} {
		in := slices.Concat(randomWalk(15, 3*period), slices.Repeat([]float64{100}, 3*period))
		in[period] *= 50 // a spike in and out
		got := apply(in, MustNewLinRegSlope(period).Process)
		for i := 4*period - 1; i < len(in); i++ {
			if g := got[i-period+1]; g != 0 {
				t.Fatalf("turns flat, period %d input %d: got %g, want 0", period, i, g)
			}
		}
	}
}

func TestRegressionInvalidPeriod(t *testing.T) {
	constructors := map[string]func(int) error{
		"LinRegSlope": func(p int) error { _, err := NewLinRegSlope(p); return err },
		"Regression":  func(p int) error { _, err := NewRegression(p); return err },
		"Correl":      func(p int) error { _, err := NewCorrel(p); return err },
	}
	for name, newX := range constructors {
		for _, p := range []int{0, -1} {
			if err := newX(p); !errors.Is(err, ErrInvalidPeriod) {
				t.Errorf("New%s(%d): got %v, want ErrInvalidPeriod", name, p, err)
			}
		}
	}
	mustConstructors := map[string]func(){
		"LinRegSlope": func() { MustNewLinRegSlope(0) },
		"Regression":  func() { MustNewRegression(0) },
		"Correl":      func() { MustNewCorrel(0) },
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

func TestRegressionNoAlloc(t *testing.T) {
	pairs := makePairs(randomWalk(1, 300), randomWalk(2, 300))
	assertNoAlloc(t, "LinRegSlope", Processor[float64, float64](MustNewLinRegSlope(14)), randomWalk(1, 300))
	assertNoAlloc(t, "Regression", Processor[Pair, RegressionValue](MustNewRegression(14)), pairs)
	assertNoAlloc(t, "Correl", Processor[Pair, float64](MustNewCorrel(14)), pairs)
}
