package ta

import "math"

// ATR computes the Average True Range over a fixed period.
type ATR struct {
	prevClose float64
	hasPrev   bool
	state     *wilderState
}

// NewATR constructs an ATR with the given period.
func NewATR(period int) (*ATR, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &ATR{state: newWilderState(period)}, nil
}

// MustNewATR constructs an ATR or panics if parameters are invalid.
func MustNewATR(period int) *ATR {
	atr, err := NewATR(period)
	if err != nil {
		panic(err)
	}
	return atr
}

// Process consumes an OHLC value and returns the ATR once enough values are available.
func (a *ATR) Process(v OHLC) (float64, bool) {
	if !a.hasPrev {
		a.prevClose = v.Close
		a.hasPrev = true
		return 0, false
	}

	tr := trueRange(v, a.prevClose)
	a.prevClose = v.Close
	return a.state.Push(tr)
}

// trueRange returns the greatest of the candle's range and its distances from
// the previous close.
func trueRange(v OHLC, prevClose float64) float64 {
	tr := v.High - v.Low
	tr = math.Max(tr, math.Abs(v.High-prevClose))
	return math.Max(tr, math.Abs(v.Low-prevClose))
}

// Reset returns the ATR to its newly constructed state.
func (a *ATR) Reset() {
	a.prevClose, a.hasPrev = 0, false
	a.state.Reset()
}
