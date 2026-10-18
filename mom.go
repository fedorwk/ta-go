package ta

// Mom computes Momentum: the difference between a value and the value a fixed
// period earlier.
type Mom struct {
	lag *lagState
}

// NewMom constructs a Momentum indicator with the given period.
func NewMom(period int) (*Mom, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Mom{lag: newLagState(period)}, nil
}

// MustNewMom constructs a Momentum indicator or panics if the period is invalid.
func MustNewMom(period int) *Mom {
	mom, err := NewMom(period)
	if err != nil {
		panic(err)
	}
	return mom
}

// Process consumes a value and returns v minus the value period inputs
// earlier, once it exists.
func (m *Mom) Process(v float64) (float64, bool) {
	old, ok := m.lag.Push(v)
	if !ok {
		return 0, false
	}
	return v - old, true
}

// Reset returns the Momentum indicator to its newly constructed state.
func (m *Mom) Reset() {
	m.lag.Reset()
}
