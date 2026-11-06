package ta

// MAD computes the median absolute deviation of the last period values.
type MAD struct {
	window *sortedWindow
}

// NewMAD constructs a MAD with the given period.
func NewMAD(period int) (*MAD, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &MAD{window: newSortedWindow(period)}, nil
}

// MustNewMAD constructs a MAD or panics if the period is invalid.
func MustNewMAD(period int) *MAD {
	m, err := NewMAD(period)
	if err != nil {
		panic(err)
	}
	return m
}

// Process consumes a value and returns median(|x[i] - m|) over the last
// period values, where m is their median (as computed by Median). The result
// is raw and unscaled; multiply by 1.4826 for an estimate of the standard
// deviation of normally distributed values. The first value is ready on the
// period-th input. A flat window, and any window of period 1, returns 0. NaN
// inputs are not supported: the result is undefined once a NaN has been
// input.
func (m *MAD) Process(v float64) (float64, bool) {
	if !m.window.Push(v) {
		return 0, false
	}
	return m.window.mad(m.window.median()), true
}

// Reset returns the MAD to its newly constructed state.
func (m *MAD) Reset() {
	m.window.Reset()
}
