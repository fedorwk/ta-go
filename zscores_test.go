package ta

import (
	"errors"
	"fmt"
	"math"
	"testing"

	talib "github.com/markcheno/go-talib"
)

// bruteRobustScore scores x against the median and MAD of base, spelling out
// the MAD = 0 and clamp rules.
func bruteRobustScore(x float64, base []float64) float64 {
	m, d := bruteMedian(base), bruteMAD(base)
	if d == 0 {
		switch {
		case x == m:
			return 0
		case x > m:
			return 10
		default:
			return -10
		}
	}
	return math.Max(-10, math.Min(10, (x-m)/(1.4826*d)))
}

// bruteRobustZ scores the last value of w against all of w.
func bruteRobustZ(w []float64) float64 {
	return bruteRobustScore(w[len(w)-1], w)
}

// bruteRobustZPrior scores the last value of w against the values before it.
func bruteRobustZPrior(w []float64) float64 {
	return bruteRobustScore(w[len(w)-1], w[:len(w)-1])
}

// bruteMeanSD returns the two-pass mean and population standard deviation
// of w.
func bruteMeanSD(w []float64) (float64, float64) {
	mean := 0.0
	for _, x := range w {
		mean += x
	}
	mean /= float64(len(w))
	m2 := 0.0
	for _, x := range w {
		m2 += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(m2 / float64(len(w)))
}

// bruteZScore scores the last value of w against the two-pass mean and
// population standard deviation of w.
func bruteZScore(w []float64) float64 {
	mean, sd := bruteMeanSD(w)
	if sd == 0 {
		return 0
	}
	return (w[len(w)-1] - mean) / sd
}

func TestRobustZCap(t *testing.T) {
	if RobustZCap != 10 {
		t.Fatalf("RobustZCap = %v, want 10 (the brute-force references assume it)", RobustZCap)
	}
}

func TestRobustZ(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewRobustZ(period).Process)
				assertAlignedApprox(t, got, bruteWindows(input, period, bruteRobustZ), period-1, 0)
			})
		}
	}
}

func TestRobustZPrior(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewRobustZPrior(period).Process)
				assertAlignedApprox(t, got, bruteWindows(input, period+1, bruteRobustZPrior), period, 0)
			})
		}
	}
}

func TestRobustZEdges(t *testing.T) {
	flat := []float64{7, 7, 7, 7, 7, 7}
	assertFloatSliceApprox(t, apply(flat, MustNewRobustZ(3).Process), []float64{0, 0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply(flat, MustNewRobustZPrior(3).Process), []float64{0, 0, 0}, 0)

	// A spike in a flat series: the MAD stays 0, so the spike scores the cap.
	spike := []float64{7, 7, 7, 7, 100, 7, 7, 7, 7, -100}
	assertFloatSliceApprox(t, apply(spike, MustNewRobustZ(5).Process), []float64{10, 0, 0, 0, 0, -10}, 0)
	assertFloatSliceApprox(t, apply(spike, MustNewRobustZPrior(4).Process), []float64{10, 0, 0, 0, 0, -10}, 0)

	// Median 3 and MAD 1: 100 is clamped, 5 scores 2/1.4826. The constant
	// 2/1.4826 is folded exactly, so allow for the runtime rounding.
	assertFloatSliceApprox(t, apply([]float64{1, 2, 3, 4, 100}, MustNewRobustZ(5).Process), []float64{10}, 0)
	assertFloatSliceApprox(t, apply([]float64{1, 2, 3, 4, 5}, MustNewRobustZ(5).Process), []float64{2 / 1.4826}, 1e-15)
	assertFloatSliceApprox(t, apply([]float64{1, 2, 3, 4, 5, 6}, MustNewRobustZPrior(5).Process), []float64{3 / 1.4826}, 1e-15)

	one := []float64{3, 3, 5, 1}
	assertFloatSliceApprox(t, apply(one, MustNewRobustZ(1).Process), []float64{0, 0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply(one, MustNewRobustZPrior(1).Process), []float64{0, 10, -10}, 0)

	// Period 2: the median is the midpoint and the MAD half the distance.
	two := []float64{1, 3, 3, 0}
	assertFloatSliceApprox(t, apply(two, MustNewRobustZ(2).Process), []float64{1 / 1.4826, 0, -1 / 1.4826}, 1e-15)

	// Median -0.1·Max and MAD 0.9·Max: the difference from the median and
	// the scale 1.4826·MAD both overflow, which would be Inf/Inf = NaN
	// without the guard.
	big := math.MaxFloat64
	assertFloatSliceApprox(t, apply([]float64{-big, -big, 0.8 * big, big}, MustNewRobustZ(4).Process), []float64{10}, 0)
	assertFloatSliceApprox(t, apply([]float64{big, big, -0.8 * big, -big}, MustNewRobustZ(4).Process), []float64{-10}, 0)
	assertFloatSliceApprox(t, apply([]float64{-big, -big, 0.8 * big, big, big}, MustNewRobustZPrior(4).Process), []float64{10}, 0)
	assertFloatSliceApprox(t, apply([]float64{big, big, -0.8 * big, -big, -big}, MustNewRobustZPrior(4).Process), []float64{-10}, 0)
}

func TestRobustZHelperFinite(t *testing.T) {
	big := math.MaxFloat64
	vals := []float64{-big, -1, -5e-324, 0, 5e-324, 1, big}
	for _, x := range vals {
		for _, med := range vals {
			for _, mad := range []float64{0, 5e-324, 1, big} {
				z := robustZ(x, med, mad)
				if math.IsNaN(z) || math.Abs(z) > RobustZCap {
					t.Fatalf("robustZ(%v, %v, %v) = %v", x, med, mad, z)
				}
			}
		}
	}
}

// maxAbs returns the largest |x| in w.
func maxAbs(w []float64) float64 {
	m := 0.0
	for _, x := range w {
		m = math.Max(m, math.Abs(x))
	}
	return m
}

// assertZScores checks got, one output per full window of input, against a
// two-pass recompute. A z-score inherits the rounding of its mean: both the
// running and the two-pass mean are off by up to about period·ε·max|x|, and
// z = (x - mean)/sd scales that by 1/sd, so the tolerance does too. Measured
// errors stay within a quarter of it.
func assertZScores(t *testing.T, got, input []float64, period int) {
	t.Helper()
	if len(got) != len(input)-period+1 {
		t.Fatalf("period %d: got %d outputs, want %d", period, len(got), len(input)-period+1)
	}
	bound := math.Sqrt(float64(period-1)) + 1e-12
	for i, g := range got {
		w := input[i : i+period]
		want := bruteZScore(w)
		_, sd := bruteMeanSD(w)
		tol := 0.0
		if sd > 0 {
			tol = 4 * float64(period+16) * 0x1p-52 * maxAbs(w) / sd
		}
		if math.Abs(g-want) > tol {
			t.Fatalf("period %d output %d: got %.12f, want %.12f (diff %g > %g)", period, i, g, want, g-want, tol)
		}
		if math.Abs(g) > bound+tol {
			t.Fatalf("period %d output %d: |%v| > √(period-1)", period, i, g)
		}
	}
}

func TestZScoreTalib(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{2, 3, 14, 50} {
		sma, sd := talib.Sma(input, period), talib.StdDev(input, period, 1)
		got := apply(input, MustNewZScore(period).Process)
		for i, g := range got {
			j := i + period - 1
			want := 0.0
			if sd[j] != 0 {
				want = (input[j] - sma[j]) / sd[j]
			}
			// go-talib computes variance as E[x^2]-E[x]^2 (see TestStdDev),
			// whose relative error grows as mean²/variance.
			mean, s := bruteMeanSD(input[i : j+1])
			tol := 64 * 0x1p-52 * mean * mean / (s * s)
			if math.Abs(g-want) > tol {
				t.Fatalf("period %d output %d: got %.12f, want %.12f (diff %g > %g)", period, i, g, want, g-want, tol)
			}
		}
	}
}

func TestZScore(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range []int{1, 2, 3, 14, 50} {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				assertZScores(t, apply(input, MustNewZScore(period).Process), input, period)
			})
		}
	}
}

func TestZScoreEdges(t *testing.T) {
	assertFloatSliceApprox(t, apply([]float64{7, 7, 7, 7, 7}, MustNewZScore(3).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply([]float64{0, 0, 0, 0}, MustNewZScore(2).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply([]float64{3, -1, 5}, MustNewZScore(1).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply([]float64{1, 3, 3, 1}, MustNewZScore(2).Process), []float64{1, 0, -1}, 0)
}

// TestZScoreTurnsFlat checks that a window that turns flat scores exactly 0,
// whether or not it does so at an exact recompute. Rounding residue in the
// running mean and variance would otherwise give noise of order 1.
func TestZScoreTurnsFlat(t *testing.T) {
	for _, offset := range []float64{0, 1e5} {
		walk := randomWalk(6, 200)
		input := make([]float64, 0, 400)
		for _, v := range walk {
			input = append(input, v+offset)
		}
		level := input[len(input)-1] + 0.1
		for range 200 {
			input = append(input, level)
		}
		for _, period := range []int{2, 3, 14, 50, 61} {
			got := apply(input, MustNewZScore(period).Process)
			// Output j is for input j+period-1, so the window is flat from
			// output len(walk) on.
			for j := len(walk); j < len(got); j++ {
				if got[j] != 0 {
					t.Fatalf("offset %g period %d: flat window output %d = %g", offset, period, j, got[j])
				}
			}
			// One more distinct value after the flat run: the window holds
			// period-1 copies of level and x, so z = √(period-1).
			p := MustNewZScore(period)
			Apply(p, input)
			z, _ := p.Process(level + 1)
			if want := math.Sqrt(float64(period - 1)); math.Abs(z-want) > 1e-9*want {
				t.Fatalf("offset %g period %d: after the flat run z = %v, want %v", offset, period, z, want)
			}
		}
	}
}

// TestZScoreSpreadCollapse feeds values that are close together near 1e5,
// with an occasional large one. When a large value leaves the window the
// variance collapses, and the residue of the running update would swamp it
// without the recompute that varianceState does then.
func TestZScoreSpreadCollapse(t *testing.T) {
	input := randomWalk(7, 20_000)
	for i := range input {
		input[i] = 1e5 + input[i]*1e-6
		if i%97 == 0 {
			input[i] += 1e3
		}
	}
	for _, period := range []int{2, 3, 14, 50, 390} {
		assertZScores(t, apply(input, MustNewZScore(period).Process), input, period)
	}
}

// TestZScoreLargeOffset checks the running mean and variance against a
// two-pass recompute over a long stream of large values.
func TestZScoreLargeOffset(t *testing.T) {
	input := randomWalk(2, 100_000)
	for i := range input {
		input[i] += 1e5
	}
	for _, period := range []int{2, 14, 50, 390} {
		assertZScores(t, apply(input, MustNewZScore(period).Process), input, period)
	}
}

func TestZScoresInvalidPeriod(t *testing.T) {
	constructors := map[string]func(int) error{
		"RobustZ":      func(p int) error { _, err := NewRobustZ(p); return err },
		"RobustZPrior": func(p int) error { _, err := NewRobustZPrior(p); return err },
		"ZScore":       func(p int) error { _, err := NewZScore(p); return err },
	}
	for name, newX := range constructors {
		for _, p := range []int{0, -1} {
			if err := newX(p); !errors.Is(err, ErrInvalidPeriod) {
				t.Errorf("New%s(%d): got %v, want ErrInvalidPeriod", name, p, err)
			}
		}
	}
	mustConstructors := map[string]func(){
		"RobustZ":      func() { MustNewRobustZ(0) },
		"RobustZPrior": func() { MustNewRobustZPrior(0) },
		"ZScore":       func() { MustNewZScore(0) },
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
