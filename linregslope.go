package ta

// LinRegSlope computes the slope of the ordinary least squares line through
// the last period values, plotted against their position in the series.
//
// It regresses each value on its arrival index with the same anchored,
// periodically recomputed sums as Regression, so it stays accurate for large
// values, such as prices near 1e5, and over long streams. The index is exact
// for the first 2⁵³ inputs.
type LinRegSlope struct {
	state *pairWindowState
	i     float64 // the arrival index of the last value
}

// NewLinRegSlope constructs a LinRegSlope with the given period.
func NewLinRegSlope(period int) (*LinRegSlope, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &LinRegSlope{state: newPairWindowState(period)}, nil
}

// MustNewLinRegSlope constructs a LinRegSlope or panics if the period is
// invalid.
func MustNewLinRegSlope(period int) *LinRegSlope {
	l, err := NewLinRegSlope(period)
	if err != nil {
		panic(err)
	}
	return l
}

// Process consumes a value and returns
//
//	Σ(i − c)(y_i − ȳ) / Σ(i − c)²,  c = (period − 1) / 2
//
// over the last period values, where y_0 is the oldest and y_{period−1} the
// current one. This is the change per input of the fitted line, as in TA-Lib's
// LINEARREG_SLOPE. The first value is ready on the period-th input. A flat
// window returns exactly 0, and so does a period of 1, which has no slope.
// Inputs must be finite.
func (l *LinRegSlope) Process(v float64) (float64, bool) {
	l.i++
	s := l.state
	if !s.push(Pair{X: l.i, Y: v}) {
		return 0, false
	}
	cxx := s.cxx()
	if cxx == 0 {
		return 0, true
	}
	return s.cxy() / cxx, true
}

// Reset returns the LinRegSlope to its newly constructed state.
func (l *LinRegSlope) Reset() {
	l.state.Reset()
	l.i = 0
}
