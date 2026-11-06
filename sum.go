package ta

// Sum computes the rolling sum of the last period values.
type Sum struct {
	state *smaState
}

// NewSum constructs a Sum with the given period.
func NewSum(period int) (*Sum, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Sum{state: newSMAState(period)}, nil
}

// MustNewSum constructs a Sum or panics if the period is invalid.
func MustNewSum(period int) *Sum {
	s, err := NewSum(period)
	if err != nil {
		panic(err)
	}
	return s
}

// Process consumes a value and returns x[t-period+1] + ... + x[t], the sum of
// the last period values. The first value is ready on the period-th input.
// There are no degenerate inputs: a window of zeros sums to 0. The running sum
// is recomputed exactly from the window every period inputs, so it does not
// drift over long streams.
func (s *Sum) Process(v float64) (float64, bool) {
	if !s.state.add(v) {
		return 0, false
	}
	return s.state.sum, true
}

// Reset returns the Sum to its newly constructed state.
func (s *Sum) Reset() {
	s.state.Reset()
}
