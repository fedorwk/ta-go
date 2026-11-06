package ta

// RelVolume compares each candle's volume with the typical volume of the
// candles before it.
type RelVolume struct {
	window *sortedWindow // the previous period volumes
}

// NewRelVolume constructs a RelVolume with the given period.
func NewRelVolume(period int) (*RelVolume, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RelVolume{window: newSortedWindow(period)}, nil
}

// MustNewRelVolume constructs a RelVolume or panics if the period is
// invalid.
func MustNewRelVolume(period int) *RelVolume {
	r, err := NewRelVolume(period)
	if err != nil {
		panic(err)
	}
	return r
}

// Process consumes an OHLCV value and returns
//
//	V / median(V[t-period], ..., V[t-1])
//
// the candle's volume relative to the median volume of the period candles
// before it, so the candle does not affect its own baseline. This is
// Chain(Lag(1), Median(period)) over volume, divided into the current
// volume. The first value is ready on candle period+1. A baseline of 0, as
// when most of the window traded nothing, returns 0. NaN inputs are not
// supported: the result is undefined once a NaN has been input.
func (r *RelVolume) Process(v OHLCV) (float64, bool) {
	if !r.window.full() {
		r.window.Push(v.Volume)
		return 0, false
	}
	base := r.window.median()
	r.window.Push(v.Volume)
	return ratioToBaseline(v.Volume, base), true
}

// Reset returns the RelVolume to its newly constructed state.
func (r *RelVolume) Reset() {
	r.window.Reset()
}
