package ta

import "math"

// TrueRange computes the true range of each candle: its high-low range,
// widened to include the previous close when the candle gapped away from it.
type TrueRange struct {
	prevClose float64
	hasPrev   bool
}

// NewTrueRange constructs a TrueRange.
func NewTrueRange() *TrueRange {
	return &TrueRange{}
}

// MustNewTrueRange constructs a TrueRange.
func MustNewTrueRange() *TrueRange {
	return NewTrueRange()
}

// Process consumes an OHLC value and returns
// max(H−L, |H−Cprev|, |L−Cprev|), where Cprev is the previous candle's close.
// The first value is ready on the 2nd candle, as with TA-Lib's TRANGE. A flat
// candle at the previous close returns 0.
func (t *TrueRange) Process(v OHLC) (float64, bool) {
	if !t.hasPrev {
		t.prevClose, t.hasPrev = v.Close, true
		return 0, false
	}
	tr := trueRange(v, t.prevClose)
	t.prevClose = v.Close
	return tr, true
}

// Reset returns the TrueRange to its newly constructed state.
func (t *TrueRange) Reset() {
	t.prevClose, t.hasPrev = 0, false
}

// trueRange returns the greatest of the candle's range and its distances from
// the previous close.
func trueRange(v OHLC, prevClose float64) float64 {
	tr := v.High - v.Low
	tr = math.Max(tr, math.Abs(v.High-prevClose))
	return math.Max(tr, math.Abs(v.Low-prevClose))
}
