package ta

// CLV computes the close location value: where each candle closed within its
// high-low range, from 0 at the low to 1 at the high.
type CLV struct{}

// NewCLV constructs a CLV.
func NewCLV() *CLV {
	return &CLV{}
}

// MustNewCLV constructs a CLV.
func MustNewCLV() *CLV {
	return NewCLV()
}

// Process consumes an OHLC value and returns
//
//	(C − L) / (H − L)
//
// in [0, 1]. Every candle is ready, the first included. A candle with
// H = L returns 0.5, and a close outside [L, H] is clamped to the range.
//
// TA-Lib uses the signed form ((C − L) − (H − C)) / (H − L), which equals
// 2·CLV − 1 and runs from −1 to 1; multiplied by volume it is one step of
// the Accumulation/Distribution line (AD).
func (c *CLV) Process(v OHLC) (float64, bool) {
	rng := v.High - v.Low
	if rng == 0 {
		return 0.5, true
	}
	return min(max((v.Close-v.Low)/rng, 0), 1), true
}

// Reset does nothing, since a CLV keeps no state; it is there so that CLV
// can be reset like every other indicator.
func (c *CLV) Reset() {}
