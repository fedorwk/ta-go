package ta

import "math"

// logReturn returns ln(v/old), or 0 unless both values are positive (NaN
// included). The ratio keeps small returns accurate; when it overflows or
// leaves the normal range, the difference of logs is used instead, which is
// finite for any finite positive values.
func logReturn(v, old float64) float64 {
	if !(v > 0 && old > 0) {
		return 0
	}
	r := v / old
	if r > math.MaxFloat64 || r < 0x1p-1022 {
		return math.Log(v) - math.Log(old)
	}
	return math.Log(r)
}

// LogReturn computes the logarithmic return over a fixed period.
type LogReturn struct {
	lag *lagState
}

// NewLogReturn constructs a LogReturn with the given period.
func NewLogReturn(period int) (*LogReturn, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &LogReturn{lag: newLagState(period)}, nil
}

// MustNewLogReturn constructs a LogReturn or panics if the period is invalid.
func MustNewLogReturn(period int) *LogReturn {
	l, err := NewLogReturn(period)
	if err != nil {
		panic(err)
	}
	return l
}

// Process consumes a value x and returns
//
//	ln(x / x[t-period])
//
// the continuously compounded counterpart of ROC. The first value is ready
// on input period+1. It is 0 when either value is ≤ 0 or NaN, and stays
// finite for finite inputs even when their ratio overflows.
func (l *LogReturn) Process(v float64) (float64, bool) {
	old, ok := l.lag.Push(v)
	if !ok {
		return 0, false
	}
	return logReturn(v, old), true
}

// Reset returns the LogReturn to its newly constructed state.
func (l *LogReturn) Reset() {
	l.lag.Reset()
}
