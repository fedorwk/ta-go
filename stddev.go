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
type varianceState struct {
	period int
	window []float64
	mean   float64
	m2     float64 // sum of squared deviations from mean over the window
	count  int
	idx    int
}

func newVarianceState(period int) *varianceState {
	return &varianceState{
		period: period,
		window: make([]float64, period),
	}
}

// Push adds a value and reports whether the window is full.
func (s *varianceState) Push(v float64) bool {
	if s.count < s.period {
		s.window[s.count] = v
		s.count++
		delta := v - s.mean
		s.mean += delta / float64(s.count)
		s.m2 += delta * (v - s.mean)
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
	if s.m2 < 0 {
		s.m2 = 0
	}
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
}

// Reset empties the window.
func (s *varianceState) Reset() {
	s.mean, s.m2, s.count, s.idx = 0, 0, 0, 0
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
