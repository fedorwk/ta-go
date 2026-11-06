package ta

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"testing"
)

var orderStatTestPeriods = []int{1, 2, 3, 20, 61}

// orderStatInputs returns a random walk, a tie-heavy series of small
// integers, and a series that is mostly zeros.
func orderStatInputs() map[string][]float64 {
	r := rand.New(rand.NewSource(3))
	ties := make([]float64, 300)
	zeros := make([]float64, 300)
	for i := range ties {
		ties[i] = float64(r.Intn(5))
		if r.Intn(4) == 0 {
			zeros[i] = r.NormFloat64()
		}
	}
	return map[string][]float64{
		"walk":  randomWalk(1, 300),
		"ties":  ties,
		"zeros": zeros,
	}
}

func sortedCopy(w []float64) []float64 {
	s := slices.Clone(w)
	slices.Sort(s)
	return s
}

func bruteMedian(w []float64) float64 {
	s := sortedCopy(w)
	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}
	return midpoint(s[n/2-1], s[n/2])
}

func bruteMAD(w []float64) float64 {
	m := bruteMedian(w)
	dev := make([]float64, len(w))
	for i, x := range w {
		if x < m {
			dev[i] = m - x
		} else {
			dev[i] = x - m
		}
	}
	return bruteMedian(dev)
}

// brutePercentRank ranks the last value of w against the others.
func brutePercentRank(w []float64) float64 {
	x, others := w[len(w)-1], w[:len(w)-1]
	if len(others) == 0 {
		return 0.5
	}
	score := 0.0
	for _, o := range others {
		switch {
		case o < x:
			score++
		case o == x:
			score += 0.5
		}
	}
	return score / float64(len(others))
}

// bruteWindows applies f to each full window of period values, aligned like
// a go-talib output (zero during warm-up).
func bruteWindows(input []float64, period int, f func([]float64) float64) []float64 {
	out := make([]float64, len(input))
	for i := period - 1; i < len(input); i++ {
		out[i] = f(input[i-period+1 : i+1])
	}
	return out
}

func TestSortedWindowInvariant(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			w := newSortedWindow(period)
			for i, v := range input {
				full := w.Push(v)
				if full != (i+1 >= period) || full != w.full() {
					t.Fatalf("%s period %d input %d: full = %v", name, period, i, full)
				}
				want := sortedCopy(w.ring[:w.count])
				if !slices.Equal(w.sorted[:w.count], want) {
					t.Fatalf("%s period %d input %d: sorted = %v, want %v", name, period, i, w.sorted[:w.count], want)
				}
			}
		}
	}
}

func TestMedian(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewMedian(period).Process)
				assertAlignedApprox(t, got, bruteWindows(input, period, bruteMedian), period-1, 0)
			})
		}
	}
}

func TestMAD(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewMAD(period).Process)
				assertAlignedApprox(t, got, bruteWindows(input, period, bruteMAD), period-1, 0)
			})
		}
	}
}

// TestMADSelection checks the binary-search selection in sortedWindow.mad at
// every small period and some large ones, on windows whose deviations are
// lopsided (a trending walk puts most values on one side of the median) or
// tied.
func TestMADSelection(t *testing.T) {
	r := rand.New(rand.NewSource(4))
	trend := make([]float64, 2000)
	ties := make([]float64, 2000)
	for i := range trend {
		trend[i] = float64(i) + r.ExpFloat64()*float64(i%7)
		ties[i] = float64(r.Intn(3))
	}
	periods := []int{4, 5, 6, 7, 8, 9, 10, 11, 390, 391}
	for name, input := range map[string][]float64{"trend": trend, "ties": ties} {
		for _, period := range periods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewMAD(period).Process)
				assertAlignedApprox(t, got, bruteWindows(input, period, bruteMAD), period-1, 0)
			})
		}
	}
}

func TestPercentRank(t *testing.T) {
	for name, input := range orderStatInputs() {
		for _, period := range orderStatTestPeriods {
			t.Run(fmt.Sprintf("%s/period=%d", name, period), func(t *testing.T) {
				got := apply(input, MustNewPercentRank(period).Process)
				assertAlignedApprox(t, got, bruteWindows(input, period, brutePercentRank), period-1, 1e-12)
			})
		}
	}
}

func TestOrderStatsEdges(t *testing.T) {
	flat := []float64{7, 7, 7, 7, 7}
	assertFloatSliceApprox(t, apply(flat, MustNewMedian(3).Process), []float64{7, 7, 7}, 0)
	assertFloatSliceApprox(t, apply(flat, MustNewMAD(3).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply(flat, MustNewPercentRank(3).Process), []float64{0.5, 0.5, 0.5}, 0)

	zeros := []float64{0, 0, 0, 0}
	assertFloatSliceApprox(t, apply(zeros, MustNewMedian(2).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply(zeros, MustNewMAD(2).Process), []float64{0, 0, 0}, 0)

	one := []float64{3, -1, 5}
	assertFloatSliceApprox(t, apply(one, MustNewMedian(1).Process), one, 0)
	assertFloatSliceApprox(t, apply(one, MustNewMAD(1).Process), []float64{0, 0, 0}, 0)
	assertFloatSliceApprox(t, apply(one, MustNewPercentRank(1).Process), []float64{0.5, 0.5, 0.5}, 0)

	// Median 3; deviations 2, 1, 0, 1, 97 have median 1; 100 is above all
	// four others.
	spike := []float64{1, 2, 3, 4, 100}
	assertFloatSliceApprox(t, apply(spike, MustNewMedian(5).Process), []float64{3}, 0)
	assertFloatSliceApprox(t, apply(spike, MustNewMAD(5).Process), []float64{1}, 0)
	assertFloatSliceApprox(t, apply(spike, MustNewPercentRank(5).Process), []float64{1}, 0)

	// Even period: median of {1, 2, 4, 10} is 3; deviations 2, 1, 1, 7 have
	// median 1.5, and 10 is above all three others. Then {2, 4, 10, 0} has median
	// 3, deviations 1, 1, 7, 3 with median 2, and 0 is below all three others.
	even := []float64{1, 2, 4, 10, 0}
	assertFloatSliceApprox(t, apply(even, MustNewMedian(4).Process), []float64{3, 3}, 0)
	assertFloatSliceApprox(t, apply(even, MustNewMAD(4).Process), []float64{1.5, 2}, 0)
	assertFloatSliceApprox(t, apply(even, MustNewPercentRank(4).Process), []float64{1, 0}, 0)
}

func TestOrderStatsInvalidPeriod(t *testing.T) {
	constructors := map[string]func(int) error{
		"Median":      func(p int) error { _, err := NewMedian(p); return err },
		"MAD":         func(p int) error { _, err := NewMAD(p); return err },
		"PercentRank": func(p int) error { _, err := NewPercentRank(p); return err },
	}
	for name, newX := range constructors {
		for _, p := range []int{0, -1} {
			if err := newX(p); !errors.Is(err, ErrInvalidPeriod) {
				t.Errorf("New%s(%d): got %v, want ErrInvalidPeriod", name, p, err)
			}
		}
	}
	mustConstructors := map[string]func(){
		"Median":      func() { MustNewMedian(0) },
		"MAD":         func() { MustNewMAD(0) },
		"PercentRank": func() { MustNewPercentRank(0) },
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
