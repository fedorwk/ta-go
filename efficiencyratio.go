package ta

import "math"

// EfficiencyRatio computes Kaufman's efficiency ratio: how much of the
// distance a series travelled over a fixed period was net movement. It is 1
// for a series that moves in one direction and near 0 for one that chops
// back and forth.
type EfficiencyRatio struct {
	start *lagState       // x[t-period]
	prev  *lagState       // x[t-1]
	path  *nonNegSumState // |x[i] - x[i-1]| over the last period steps
}

// NewEfficiencyRatio constructs an EfficiencyRatio with the given period.
func NewEfficiencyRatio(period int) (*EfficiencyRatio, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &EfficiencyRatio{
		start: newLagState(period),
		prev:  newLagState(1),
		path:  newNonNegSumState(period),
	}, nil
}

// MustNewEfficiencyRatio constructs an EfficiencyRatio or panics if the
// period is invalid.
func MustNewEfficiencyRatio(period int) *EfficiencyRatio {
	e, err := NewEfficiencyRatio(period)
	if err != nil {
		panic(err)
	}
	return e
}

// Process consumes a value x and returns
//
//	|x − x[t-period]| / (|x[t-period+1] − x[t-period]| + ... + |x − x[t-1]|)
//
// the net change over the last period steps divided by the path length of
// those steps, in [0, 1]. The first value is ready on input period+1. A flat
// window, whose path length is 0, returns 0, and a period of 1 returns 1 for
// any change and 0 for none. The result is clamped to 1 to absorb rounding.
// Inputs must be finite.
func (e *EfficiencyRatio) Process(v float64) (float64, bool) {
	start, _ := e.start.Push(v)
	prev, ok := e.prev.Push(v)
	if !ok {
		return 0, false
	}
	// The path window fills on input period+1, the same input as start.
	if !e.path.add(math.Abs(v - prev)) {
		return 0, false
	}
	path := e.path.sum()
	if path == 0 {
		return 0, true
	}
	return min(math.Abs(v-start)/path, 1), true
}

// Reset returns the EfficiencyRatio to its newly constructed state.
func (e *EfficiencyRatio) Reset() {
	e.start.Reset()
	e.prev.Reset()
	e.path.Reset()
}
