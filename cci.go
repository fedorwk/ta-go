package ta

import "math"

// CCI computes the Commodity Channel Index over a fixed period.
type CCI struct {
	period int
	window []float64 // typical prices
	count  int
	idx    int
}

// NewCCI constructs a CCI with the given period.
func NewCCI(period int) (*CCI, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &CCI{period: period, window: make([]float64, period)}, nil
}

// MustNewCCI constructs a CCI or panics if the period is invalid.
func MustNewCCI(period int) *CCI {
	cci, err := NewCCI(period)
	if err != nil {
		panic(err)
	}
	return cci
}

// Process consumes an OHLC value and returns the CCI once enough values are
// available. It is 0 when the typical prices in the window are all equal.
//
// Each value costs O(period): the mean absolute deviation has no rolling form.
func (c *CCI) Process(v OHLC) (float64, bool) {
	tp := (v.High + v.Low + v.Close) / 3
	c.window[c.idx] = tp
	c.idx++
	if c.idx == c.period {
		c.idx = 0
	}
	if c.count < c.period {
		c.count++
		if c.count < c.period {
			return 0, false
		}
	}

	mean := 0.0
	for _, x := range c.window {
		mean += x
	}
	mean /= float64(c.period)
	meanDev := 0.0
	for _, x := range c.window {
		meanDev += math.Abs(x - mean)
	}
	meanDev /= float64(c.period)

	if meanDev == 0 {
		return 0, true
	}
	return (tp - mean) / (0.015 * meanDev), true
}

// Reset returns the CCI to its newly constructed state.
func (c *CCI) Reset() {
	c.count, c.idx = 0, 0
}
