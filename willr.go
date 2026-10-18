package ta

// WillR computes Williams %R over a fixed period.
type WillR struct {
	highest *extremumState
	lowest  *extremumState
}

// NewWillR constructs a Williams %R with the given period.
func NewWillR(period int) (*WillR, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &WillR{highest: newMaxState(period), lowest: newMinState(period)}, nil
}

// MustNewWillR constructs a Williams %R or panics if the period is invalid.
func MustNewWillR(period int) *WillR {
	w, err := NewWillR(period)
	if err != nil {
		panic(err)
	}
	return w
}

// Process consumes an OHLC value and returns Williams %R, in [-100, 0], once
// enough values are available. It is 0 when the window's high equals its low.
func (w *WillR) Process(v OHLC) (float64, bool) {
	highest, ready := w.highest.Push(v.High)
	lowest, _ := w.lowest.Push(v.Low)
	if !ready {
		return 0, false
	}
	denom := highest - lowest
	if denom == 0 {
		return 0, true
	}
	return -100 * (highest - v.Close) / denom, true
}

// Reset returns the Williams %R to its newly constructed state.
func (w *WillR) Reset() {
	w.highest.Reset()
	w.lowest.Reset()
}
