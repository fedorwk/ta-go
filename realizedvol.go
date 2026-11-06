package ta

import "math"

// RealizedVol computes realized volatility: the root of the sum of squared
// one-step log returns over a fixed period. The squared returns are summed by
// nonNegSumState, so the sum neither drifts over long streams nor keeps the
// residue of large returns that have left the window.
type RealizedVol struct {
	lag *lagState       // the previous value
	sq  *nonNegSumState // squared one-step log returns
}

// NewRealizedVol constructs a RealizedVol with the given period.
func NewRealizedVol(period int) (*RealizedVol, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RealizedVol{lag: newLagState(1), sq: newNonNegSumState(period)}, nil
}

// MustNewRealizedVol constructs a RealizedVol or panics if the period is
// invalid.
func MustNewRealizedVol(period int) *RealizedVol {
	r, err := NewRealizedVol(period)
	if err != nil {
		panic(err)
	}
	return r
}

// Process consumes a value and returns
//
//	√(r[t-period+1]² + ... + r[t]²),  r[i] = ln(x[i] / x[i-1])
//
// over the last period one-step log returns. The first value is ready on
// input period+1. It is neither annualised nor de-meaned: it is the root of
// the sum, not a standard deviation, so scale it by √(periods per year /
// period) to annualise. A return that involves a value ≤ 0 counts as 0, as
// in LogReturn, and a flat window returns exactly 0.
func (r *RealizedVol) Process(v float64) (float64, bool) {
	prev, ok := r.lag.Push(v)
	if !ok {
		return 0, false
	}
	lr := logReturn(v, prev)
	if !r.sq.add(lr * lr) {
		return 0, false
	}
	return math.Sqrt(r.sq.sum()), true
}

// Reset returns the RealizedVol to its newly constructed state.
func (r *RealizedVol) Reset() {
	r.lag.Reset()
	r.sq.Reset()
}
