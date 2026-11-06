package ta

import "sort"

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

// mad returns the median of |x - m| over the window. It walks outward from
// m, taking the smaller deviation on either side at each step, so the
// deviations come out in ascending order and the walk stops after about n/2
// steps. m must lie within the window's range, as the median does. The
// window must not be empty.
func (w *sortedWindow) mad(m float64) float64 {
	s, n := w.sorted[:w.count], w.count
	r := sort.SearchFloat64s(s, m)
	l := r - 1
	var prev, cur float64
	for k := 0; k <= n/2; k++ {
		prev = cur
		if l >= 0 && (r >= n || m-s[l] <= s[r]-m) {
			cur = m - s[l]
			l--
		} else {
			cur = s[r] - m
			r++
		}
	}
	if n%2 == 1 {
		return cur
	}
	return midpoint(prev, cur)
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

// midpoint returns (a+b)/2 without overflowing for large finite a and b.
// Halving is exact for normal floats, so this rounds the same as (a+b)/2.
func midpoint(a, b float64) float64 {
	return a/2 + b/2
}
