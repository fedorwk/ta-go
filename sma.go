package ta

// SMA computes the Simple Moving Average over a fixed period.
type SMA struct {
	state *smaState
}

// smaState maintains a rolling simple moving average.
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
	if s.count < s.period {
		s.window[s.count] = v
		s.sum += v
		s.count++
		if s.count == s.period {
			return s.sum / float64(s.period), true
		}
		return 0, false
	}

	s.sum -= s.window[s.idx]
	s.window[s.idx] = v
	s.sum += v
	s.idx++
	if s.idx == s.period {
		s.idx = 0
	}

	return s.sum / float64(s.period), true
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
