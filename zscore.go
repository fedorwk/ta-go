package ta

// ZScore scores the current value against the mean and standard deviation of
// the last period values, including itself.
type ZScore struct {
	state *varianceState
}

// NewZScore constructs a ZScore with the given period.
func NewZScore(period int) (*ZScore, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &ZScore{state: newVarianceState(period)}, nil
}

// MustNewZScore constructs a ZScore or panics if the period is invalid.
func MustNewZScore(period int) *ZScore {
	z, err := NewZScore(period)
	if err != nil {
		panic(err)
	}
	return z
}

// Process consumes a value x and returns
//
//	(x - mean) / stddev
//
// over the last period values, x included, with the population standard
// deviation as computed by StdDev. The first value is ready on the period-th
// input. Since x is in the window, |z| <= √(period-1). A window whose
// standard deviation is 0 returns 0, so a flat window and a period of 1
// always return 0. For a value scored against a robust baseline, see
// RobustZ and RobustZPrior.
func (z *ZScore) Process(v float64) (float64, bool) {
	if !z.state.Push(v) {
		return 0, false
	}
	sd := z.state.stdDev()
	if sd == 0 {
		return 0, true
	}
	return (v - z.state.mean) / sd, true
}

// Reset returns the ZScore to its newly constructed state.
func (z *ZScore) Reset() {
	z.state.Reset()
}
