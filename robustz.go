package ta

import "math"

// RobustZCap bounds the output of RobustZ and RobustZPrior to
// [-RobustZCap, RobustZCap]. Without it, one value far outside its window
// could score an arbitrarily large z, and any value off the median of a
// window whose MAD is 0 (more than half the values equal) would score an
// infinite one. At the cap, a single outlier cannot dominate a downstream
// sum, average or threshold.
const RobustZCap = 10

// madNormalScale turns a MAD into an estimate of the standard deviation of
// normally distributed values.
const madNormalScale = 1.4826

// robustZ returns (x - med) / (madNormalScale·mad) clamped to ±RobustZCap. A
// mad of 0 gives 0 when x equals med and ±RobustZCap otherwise, since a
// non-zero difference divided by 0 is ±Inf.
func robustZ(x, med, mad float64) float64 {
	d := x - med
	if d == 0 {
		return 0
	}
	if math.IsInf(d, 0) {
		// The difference overflowed. The scale may have too, and Inf/Inf
		// would be NaN.
		return math.Copysign(RobustZCap, d)
	}
	return max(-RobustZCap, min(RobustZCap, d/(madNormalScale*mad)))
}

// RobustZ scores the current value against the median and MAD of the last
// period values, including itself.
type RobustZ struct {
	window *sortedWindow
}

// NewRobustZ constructs a RobustZ with the given period.
func NewRobustZ(period int) (*RobustZ, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RobustZ{window: newSortedWindow(period)}, nil
}

// MustNewRobustZ constructs a RobustZ or panics if the period is invalid.
func MustNewRobustZ(period int) *RobustZ {
	r, err := NewRobustZ(period)
	if err != nil {
		panic(err)
	}
	return r
}

// Process consumes a value x and returns
//
//	(x - median) / (1.4826·MAD)
//
// clamped to ±RobustZCap, where the median and MAD (as computed by Median
// and MAD) are over the last period values, x included. The first value is
// ready on the period-th input.
//
// Because x is part of its own baseline, a spike pulls the median and MAD
// towards itself, though only slightly, as both are robust to a single
// outlier. Use RobustZPrior to score x against only the values before it.
//
// When the MAD is 0 the result is 0 if x equals the median and ±RobustZCap
// otherwise. So a flat window returns 0, a period of 1 always returns 0, and
// a period of 2 returns 0 or ±1/1.4826. NaN inputs are not supported: the
// result is undefined once a NaN has been input.
func (r *RobustZ) Process(v float64) (float64, bool) {
	if !r.window.Push(v) {
		return 0, false
	}
	med := r.window.median()
	return robustZ(v, med, r.window.mad(med)), true
}

// Reset returns the RobustZ to its newly constructed state.
func (r *RobustZ) Reset() {
	r.window.Reset()
}

// RobustZPrior scores the current value against the median and MAD of the
// period values before it.
type RobustZPrior struct {
	window *sortedWindow
}

// NewRobustZPrior constructs a RobustZPrior with the given period.
func NewRobustZPrior(period int) (*RobustZPrior, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RobustZPrior{window: newSortedWindow(period)}, nil
}

// MustNewRobustZPrior constructs a RobustZPrior or panics if the period is
// invalid.
func MustNewRobustZPrior(period int) *RobustZPrior {
	r, err := NewRobustZPrior(period)
	if err != nil {
		panic(err)
	}
	return r
}

// Process consumes a value x and returns
//
//	(x - median) / (1.4826·MAD)
//
// clamped to ±RobustZCap, where the median and MAD are over the period
// values before x, so x does not affect its own baseline. The first value is
// ready on input period+1.
//
// When the MAD is 0 the result is 0 if x equals the median and ±RobustZCap
// otherwise. So x scores 0 against a flat window equal to it and ±RobustZCap
// against any other flat window, and a period of 1 returns 0 when x equals
// the value before it and ±RobustZCap otherwise. NaN inputs are not
// supported: the result is undefined once a NaN has been input.
func (r *RobustZPrior) Process(v float64) (float64, bool) {
	if !r.window.full() {
		r.window.Push(v)
		return 0, false
	}
	med := r.window.median()
	z := robustZ(v, med, r.window.mad(med))
	r.window.Push(v)
	return z, true
}

// Reset returns the RobustZPrior to its newly constructed state.
func (r *RobustZPrior) Reset() {
	r.window.Reset()
}
