package ta

import "math"

// RangeExpansion compares each candle's true range with the typical true
// range of the candles before it.
type RangeExpansion struct {
	tr     TrueRange
	window *sortedWindow // the previous period true ranges
}

// NewRangeExpansion constructs a RangeExpansion with the given period.
func NewRangeExpansion(period int) (*RangeExpansion, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RangeExpansion{window: newSortedWindow(period)}, nil
}

// MustNewRangeExpansion constructs a RangeExpansion or panics if the period
// is invalid.
func MustNewRangeExpansion(period int) *RangeExpansion {
	r, err := NewRangeExpansion(period)
	if err != nil {
		panic(err)
	}
	return r
}

// Process consumes an OHLC value and returns
//
//	TR / median(TR[t-period], ..., TR[t-1])
//
// the candle's true range relative to the median true range of the period
// candles before it, so the candle does not affect its own baseline. This is
// Chain(Lag(1), Median(period)) over TrueRange, divided into the current
// true range. The first true range comes on the 2nd candle, so the first
// value is ready on candle period+2. A baseline of 0, as in a flat market,
// returns 0. NaN inputs are not supported: the result is undefined once a
// NaN has been input.
func (r *RangeExpansion) Process(v OHLC) (float64, bool) {
	tr, ok := r.tr.Process(v)
	if !ok {
		return 0, false
	}
	if !r.window.full() {
		r.window.Push(tr)
		return 0, false
	}
	base := r.window.median()
	r.window.Push(tr)
	return ratioToBaseline(tr, base), true
}

// Reset returns the RangeExpansion to its newly constructed state.
func (r *RangeExpansion) Reset() {
	r.tr.Reset()
	r.window.Reset()
}

// ratioToBaseline returns x / base, or 0 when base ≤ 0. The result is
// clamped to ±math.MaxFloat64, so a tiny baseline cannot make it infinite.
func ratioToBaseline(x, base float64) float64 {
	if !(base > 0) {
		return 0
	}
	return max(min(x/base, math.MaxFloat64), -math.MaxFloat64)
}
