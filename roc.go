package ta

// ROC computes the Rate of Change, in percent, over a fixed period.
type ROC struct {
	lag *lagState
}

// NewROC constructs a ROC with the given period.
func NewROC(period int) (*ROC, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &ROC{lag: newLagState(period)}, nil
}

// MustNewROC constructs a ROC or panics if the period is invalid.
func MustNewROC(period int) *ROC {
	roc, err := NewROC(period)
	if err != nil {
		panic(err)
	}
	return roc
}

// Process consumes a value and returns 100*(v/old - 1), where old is the
// value period inputs earlier, once it exists. It is 0 when old is 0.
func (r *ROC) Process(v float64) (float64, bool) {
	old, ok := r.lag.Push(v)
	if !ok {
		return 0, false
	}
	if old == 0 {
		return 0, true
	}
	return (v/old - 1) * 100, true
}

// Reset returns the ROC to its newly constructed state.
func (r *ROC) Reset() {
	r.lag.Reset()
}
