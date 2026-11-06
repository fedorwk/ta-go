package ta

import "math"

// RealizedVol computes realized volatility: the root of the sum of squared
// one-step log returns over a fixed period.
//
// The squared returns are summed by smaState, which recomputes the sum
// exactly each time its ring wraps. Between wraps, each update rounds the
// sum by about one ulp of its current size, so when a large return leaves
// the window the residue of the larger sum can swamp the new one. As in
// varianceState, the sum is therefore also recomputed whenever it falls
// below m2DropLimit of its peak since the last exact computation. That needs
// the sum to collapse by that factor, so it is rare and the cost stays
// amortized O(1).
type RealizedVol struct {
	lag  *lagState // the previous value
	sq   *smaState // squared one-step log returns
	peak float64   // the largest sq.sum since its last exact computation
}

// NewRealizedVol constructs a RealizedVol with the given period.
func NewRealizedVol(period int) (*RealizedVol, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RealizedVol{lag: newLagState(1), sq: newSMAState(period)}, nil
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
	switch {
	case r.sq.idx == 0: // the window just filled, or wrapped and was recomputed
		r.peak = r.sq.sum
	case r.sq.sum < r.peak*m2DropLimit:
		r.sq.recompute()
		r.peak = r.sq.sum
	default:
		r.peak = max(r.peak, r.sq.sum)
	}
	return math.Sqrt(max(r.sq.sum, 0)), true
}

// Reset returns the RealizedVol to its newly constructed state.
func (r *RealizedVol) Reset() {
	r.lag.Reset()
	r.sq.Reset()
	r.peak = 0
}
