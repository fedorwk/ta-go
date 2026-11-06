package ta

import (
	"math"
	"sort"
)

// sortedWindow holds the last period values both in arrival order and sorted,
// so that order statistics of the window are available after each value.
//
// Each Push finds the evicted and inserted positions by binary search and
// moves the values between them with a single copy: O(log n) search plus an
// O(n) memmove in the worst case, with no allocation.
//
// NaN values are not supported: they break the sorted order, and the results
// are undefined from then on.
type sortedWindow struct {
	period int
	ring   []float64 // values in arrival order
	sorted []float64 // the same values in ascending order
	count  int
	idx    int // next position in ring to overwrite once full
}

func newSortedWindow(period int) *sortedWindow {
	return &sortedWindow{
		period: period,
		ring:   make([]float64, period),
		sorted: make([]float64, period),
	}
}

// Push adds a value, evicting the oldest once the window is full, and reports
// whether the window is full.
func (w *sortedWindow) Push(v float64) bool {
	if w.count < w.period {
		w.ring[w.count] = v
		j := sort.SearchFloat64s(w.sorted[:w.count], v)
		copy(w.sorted[j+1:w.count+1], w.sorted[j:w.count])
		w.sorted[j] = v
		w.count++
		return w.count == w.period
	}

	old := w.ring[w.idx]
	w.ring[w.idx] = v
	w.idx++
	if w.idx == w.period {
		w.idx = 0
	}

	// old is in sorted, so sorted[i] == old. Replace it with v by shifting
	// the values between the two positions one place towards i.
	i := sort.SearchFloat64s(w.sorted, old)
	j := sort.SearchFloat64s(w.sorted, v)
	if j <= i {
		copy(w.sorted[j+1:i+1], w.sorted[j:i])
		w.sorted[j] = v
	} else {
		copy(w.sorted[i:j-1], w.sorted[i+1:j])
		w.sorted[j-1] = v
	}
	return true
}

// full reports whether the window holds period values.
func (w *sortedWindow) full() bool {
	return w.count == w.period
}

// Reset empties the window.
func (w *sortedWindow) Reset() {
	w.count, w.idx = 0, 0
}

// median returns the median of the window; for an even count, the mean of
// the two middle values. The window must not be empty.
func (w *sortedWindow) median() float64 {
	n := w.count
	if n%2 == 1 {
		return w.sorted[n/2]
	}
	return midpoint(w.sorted[n/2-1], w.sorted[n/2])
}

// mad returns the median of |x - m| over the window in O(log n). m need
// not be the window's median or lie within its range: any m for which the
// deviations are finite works. The window must not be empty.
//
// Splitting the sorted window at m gives two ascending runs of deviations:
//
//	below(i) = m - s[r-1-i]  for i in [0, r)    (values below m, nearest first)
//	above(i) = s[r+i] - m    for i in [0, n-r)  (values at or above m)
//
// The k = n/2 smallest deviations are the first a of one run and the first
// k-a of the other, and a is found by binary search, as when selecting the
// k-th smallest of two sorted arrays. The next deviation after them is the
// one at rank n/2, and for an even n the largest of them is at rank n/2-1.
func (w *sortedWindow) mad(m float64) float64 {
	s, n := w.sorted[:w.count], w.count
	r := sort.SearchFloat64s(s, m) // s[:r] < m <= s[r:]
	nAbove := n - r
	k := n / 2

	// Find the smallest a for which taking a from below and k-a from above
	// is not too few from below, that is, below(a) >= above(k-a-1). Inside
	// the loop both indexes are in range.
	lo, hi := max(0, k-nAbove), min(k, r)
	for lo < hi {
		a := (lo + hi) / 2
		if m-s[r-1-a] < s[r+k-a-1]-m {
			lo = a + 1
		} else {
			hi = a
		}
	}
	a, b := lo, k-lo

	// The deviation at rank n/2: the smaller of the next one in each run.
	// a+b < n, so at least one run has one left.
	next := math.Inf(1)
	if a < r {
		next = m - s[r-1-a]
	}
	if b < nAbove {
		next = min(next, s[r+b]-m)
	}
	if n%2 == 1 {
		return next
	}

	// The deviation at rank n/2-1: the larger of the last one taken from
	// each run. k >= 1 here, so at least one run had one taken.
	last := math.Inf(-1)
	if a > 0 {
		last = m - s[r-a]
	}
	if b > 0 {
		last = max(last, s[r+b-1]-m)
	}
	return midpoint(last, next)
}

// countBelow returns the number of values in the window less than x.
func (w *sortedWindow) countBelow(x float64) int {
	return sort.SearchFloat64s(w.sorted[:w.count], x)
}

// countEqual returns the number of values in the window equal to x.
func (w *sortedWindow) countEqual(x float64) int {
	s := w.sorted[:w.count]
	above := sort.Search(len(s), func(i int) bool { return s[i] > x })
	return above - sort.SearchFloat64s(s, x)
}

// midpoint returns (a+b)/2, correctly rounded, for finite a and b. Only when
// a+b overflows does it halve first; both values are then far above the
// subnormal range, so halving them is exact.
func midpoint(a, b float64) float64 {
	if s := a + b; !math.IsInf(s, 0) {
		return s / 2
	}
	return a/2 + b/2
}
