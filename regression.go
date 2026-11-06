package ta

// Regression fits Y on X by ordinary least squares over the last period
// pairs.
//
// This is a plain OLS fit of the raw values. It is not TA-Lib's BETA, which
// regresses the percentage changes of two price series.
type Regression struct {
	state *pairWindowState
}

// NewRegression constructs a Regression with the given period.
func NewRegression(period int) (*Regression, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Regression{state: newPairWindowState(period)}, nil
}

// MustNewRegression constructs a Regression or panics if the period is
// invalid.
func MustNewRegression(period int) *Regression {
	r, err := NewRegression(period)
	if err != nil {
		panic(err)
	}
	return r
}

// Process consumes a pair and returns, over the last period pairs,
//
//	Slope     = Σ(x − x̄)(y − ȳ) / Σ(x − x̄)²
//	Intercept = ȳ − Slope·x̄
//	R2        = (Σ(x − x̄)(y − ȳ))² / (Σ(x − x̄)² · Σ(y − ȳ)²)
//
// The first value is ready on the period-th pair. When every X in the window
// is equal, the slope is undefined and the result is {0, ȳ, 0}, so a period
// of 1 always returns {0, y, 0}. When every Y in the window is equal, the
// slope and R2 are 0 and the intercept is that Y. R2 is clamped to [0, 1]
// against rounding. Inputs must be finite.
func (r *Regression) Process(p Pair) (RegressionValue, bool) {
	s := r.state
	if !s.push(p) {
		return RegressionValue{}, false
	}
	cxx := s.cxx()
	if cxx == 0 {
		return RegressionValue{Intercept: s.meanY()}, true
	}
	cxy := s.cxy()
	slope := cxy / cxx
	v := RegressionValue{Slope: slope, Intercept: s.meanY() - slope*s.meanX()}
	if cyy := s.cyy(); cyy > 0 {
		v.R2 = min(slope*cxy/cyy, 1)
	}
	return v, true
}

// Reset returns the Regression to its newly constructed state.
func (r *Regression) Reset() {
	r.state.Reset()
}
