package ta

import "math"

// Correl computes the Pearson correlation of X and Y over the last period
// pairs.
type Correl struct {
	state *pairWindowState
}

// NewCorrel constructs a Correl with the given period.
func NewCorrel(period int) (*Correl, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Correl{state: newPairWindowState(period)}, nil
}

// MustNewCorrel constructs a Correl or panics if the period is invalid.
func MustNewCorrel(period int) *Correl {
	c, err := NewCorrel(period)
	if err != nil {
		panic(err)
	}
	return c
}

// Process consumes a pair and returns
//
//	Σ(x − x̄)(y − ȳ) / √(Σ(x − x̄)² · Σ(y − ȳ)²)
//
// over the last period pairs, in [−1, 1]. The first value is ready on the
// period-th pair. When every X or every Y in the window is equal, the
// correlation is undefined and the result is 0, so a period of 1 always
// returns 0. The result is clamped to [−1, 1] against rounding. Inputs must
// be finite.
func (c *Correl) Process(p Pair) (float64, bool) {
	s := c.state
	if !s.push(p) {
		return 0, false
	}
	cxx, cyy := s.cxx(), s.cyy()
	if cxx == 0 || cyy == 0 {
		return 0, true
	}
	r := s.cxy() / (math.Sqrt(cxx) * math.Sqrt(cyy))
	return max(-1, min(r, 1)), true
}

// Reset returns the Correl to its newly constructed state.
func (c *Correl) Reset() {
	c.state.Reset()
}
