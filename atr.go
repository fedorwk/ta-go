package ta

import "math"

// ATR computes the Average True Range over a fixed period.
type ATR struct {
	period    int
	prevClose float64
	hasPrev   bool
	trSum     float64
	count     int
	atr       float64
	ready     bool
}

// NewATR constructs an ATR with the given period.
func NewATR(period int) (*ATR, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &ATR{period: period}, nil
}

// MustNewATR constructs an ATR or panics if parameters are invalid.
func MustNewATR(period int) *ATR {
	atr, err := NewATR(period)
	if err != nil {
		panic(err)
	}
	return atr
}

// Update consumes an OHLC value and returns the ATR once enough values are available.
func (a *ATR) Process(v OHLC) (float64, bool) {
	if !a.hasPrev {
		a.prevClose = v.Close
		a.hasPrev = true
		return 0, false
	}

	tr := v.High - v.Low
	tr = math.Max(tr, math.Abs(v.High-a.prevClose))
	tr = math.Max(tr, math.Abs(v.Low-a.prevClose))
	a.prevClose = v.Close

	if !a.ready {
		a.trSum += tr
		a.count++
		if a.count == a.period {
			a.atr = a.trSum / float64(a.period)
			a.ready = true
			return a.atr, true
		}
		return 0, false
	}

	a.atr = (a.atr*float64(a.period-1) + tr) / float64(a.period)
	return a.atr, true
}
