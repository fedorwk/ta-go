package ta

import "math"

// StdDev computes the population standard deviation over a fixed period.
type StdDev struct {
	state *varianceState
}

// varianceState maintains the mean and sum of squared deviations of a sliding
// window with a sliding-window form of Welford's algorithm, which stays
// accurate when the values are large relative to their spread. Both are
// recomputed exactly from the window each time it wraps, so rounding error
// cannot accumulate over long streams.
//
// Each update rounds m2 by about one ulp of its current size, so when the
// spread collapses (a large value leaves a window of close values) the
// residue left from the larger m2 can swamp the new one. Both are therefore
// also recomputed whenever m2 falls below m2DropLimit of its peak since the
// last recompute, which bounds the relative error of m2 by about
// period·ε/m2DropLimit. Every such recompute needs the spread to have
// collapsed by that factor, so it is rare and the cost stays amortized O(1)
// on market data.
type varianceState struct {
	period int
	window []float64
	mean   float64
	m2     float64 // sum of squared deviations from mean over the window
	count  int
	idx    int
	peak   float64 // the largest m2 since the last exact computation
	last   float64 // the last value pushed
	run    int     // trailing count of values equal to last, at most period
}

// m2DropLimit is the fraction of its peak that m2 may fall to before
// varianceState recomputes it from the window.
const m2DropLimit = 0x1p-10

func newVarianceState(period int) *varianceState {
	return &varianceState{
		period: period,
		window: make([]float64, period),
	}
}

// Push adds a value and reports whether the window is full.
//
// A window of equal values gets mean exactly v and m2 exactly 0. Rounding
// alone would leave residue in both, of comparable size, so their ratio, a
// z-score, would be noise of order 1 rather than 0. A sum of equal values is
// not exact either, so even the two-pass recompute has this problem.
func (s *varianceState) Push(v float64) bool {
	if s.run > 0 && v == s.last {
		s.run = min(s.run+1, s.period)
	} else {
		s.run = 1
	}
	s.last = v
	if !s.push(v) {
		return false
	}
	if s.run == s.period {
		s.mean, s.m2, s.peak = v, 0, 0
	}
	return true
}

// push adds a value to the window and updates mean and m2.
func (s *varianceState) push(v float64) bool {
	if s.count < s.period {
		s.window[s.count] = v
		s.count++
		delta := v - s.mean
		s.mean += delta / float64(s.count)
		s.m2 += delta * (v - s.mean) // Welford: m2 only grows here
		s.peak = s.m2
		return s.count == s.period
	}

	old := s.window[s.idx]
	s.window[s.idx] = v
	s.idx++
	if s.idx == s.period {
		s.idx = 0
		s.recompute()
		return true
	}

	oldMean := s.mean
	s.mean += (v - old) / float64(s.period)
	s.m2 += (v - old) * (v - s.mean + old - oldMean)
	if s.m2 < s.peak*m2DropLimit {
		s.recompute()
		return true
	}
	if s.m2 < 0 {
		s.m2 = 0
	}
	s.peak = max(s.peak, s.m2)
	return true
}

// recompute sets mean and m2 from the window with a two-pass computation.
func (s *varianceState) recompute() {
	sum := 0.0
	for _, x := range s.window {
		sum += x
	}
	s.mean = sum / float64(s.period)
	s.m2 = 0
	for _, x := range s.window {
		d := x - s.mean
		s.m2 += d * d
	}
	s.peak = s.m2
}

// Reset empties the window.
func (s *varianceState) Reset() {
	s.mean, s.m2, s.peak, s.count, s.idx, s.run = 0, 0, 0, 0, 0, 0
}

// stdDev returns the population standard deviation of the window.
func (s *varianceState) stdDev() float64 {
	return math.Sqrt(s.m2 / float64(s.period))
}

// NewStdDev constructs a StdDev with the given period.
func NewStdDev(period int) (*StdDev, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &StdDev{state: newVarianceState(period)}, nil
}

// MustNewStdDev constructs a StdDev or panics if the period is invalid.
func MustNewStdDev(period int) *StdDev {
	sd, err := NewStdDev(period)
	if err != nil {
		panic(err)
	}
	return sd
}

// Process consumes a value and returns the standard deviation once enough
// values are available.
func (s *StdDev) Process(v float64) (float64, bool) {
	if !s.state.Push(v) {
		return 0, false
	}
	return s.state.stdDev(), true
}

// Reset returns the StdDev to its newly constructed state.
func (s *StdDev) Reset() {
	s.state.Reset()
}
