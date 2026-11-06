package ta

// Median computes the median of the last period values.
type Median struct {
	window *sortedWindow
}

// NewMedian constructs a Median with the given period.
func NewMedian(period int) (*Median, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Median{window: newSortedWindow(period)}, nil
}

// MustNewMedian constructs a Median or panics if the period is invalid.
func MustNewMedian(period int) *Median {
	m, err := NewMedian(period)
	if err != nil {
		panic(err)
	}
	return m
}

// Process consumes a value and returns the median of the last period values:
// the middle value for an odd period, and the mean of the two middle values
// for an even period. The first value is ready on the period-th input. A flat
// window returns its value. NaN inputs are not supported: the result is
// undefined once a NaN has been input.
func (m *Median) Process(v float64) (float64, bool) {
	if !m.window.Push(v) {
		return 0, false
	}
	return m.window.median(), true
}

// Reset returns the Median to its newly constructed state.
func (m *Median) Reset() {
	m.window.Reset()
}
