package ta

// SMA computes the Simple Moving Average over a fixed period.
type SMA struct {
	state *smaState
}

// smaState maintains the sum and mean of a sliding window. The running sum is
// recomputed exactly from the window each time the ring wraps, so rounding
// error cannot accumulate over long streams; the cost stays amortized O(1).
type smaState struct {
	period int
	window []float64
	sum    float64
	count  int
	idx    int
}

func newSMAState(period int) *smaState {
	return &smaState{
		period: period,
		window: make([]float64, period),
	}
}

// Push adds a value and returns the SMA when ready.
func (s *smaState) Push(v float64) (float64, bool) {
	if !s.add(v) {
		return 0, false
	}
	return s.sum / float64(s.period), true
}

// add adds a value to the window and reports whether the window is full, in
// which case sum holds the sum of the last period values.
func (s *smaState) add(v float64) bool {
	if s.count < s.period {
		s.window[s.count] = v
		s.sum += v
		s.count++
		return s.count == s.period
	}

	s.sum -= s.window[s.idx]
	s.window[s.idx] = v
	s.sum += v
	s.idx++
	if s.idx == s.period {
		s.idx = 0
		s.recompute()
	}
	return true
}

// recompute sets sum exactly from the window.
func (s *smaState) recompute() {
	sum := 0.0
	for _, x := range s.window {
		sum += x
	}
	s.sum = sum
}

// Reset empties the window.
func (s *smaState) Reset() {
	s.sum, s.count, s.idx = 0, 0, 0
}

// NewSMA constructs an SMA with the given period.
func NewSMA(period int) (*SMA, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &SMA{state: newSMAState(period)}, nil
}

// MustNewSMA constructs an SMA or panics if the period is invalid.
func MustNewSMA(period int) *SMA {
	sma, err := NewSMA(period)
	if err != nil {
		panic(err)
	}
	return sma
}

// Process consumes a value and returns the SMA once enough values are available.
func (s *SMA) Process(v float64) (float64, bool) {
	return s.state.Push(v)
}

// Reset returns the SMA to its newly constructed state.
func (s *SMA) Reset() {
	s.state.Reset()
}
