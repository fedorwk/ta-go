package ta

// PercentRank computes where the current value ranks among the other values
// in its window.
type PercentRank struct {
	window *sortedWindow
}

// NewPercentRank constructs a PercentRank with the given period. The window
// includes the current value, so it is compared against the period-1 values
// before it.
func NewPercentRank(period int) (*PercentRank, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &PercentRank{window: newSortedWindow(period)}, nil
}

// MustNewPercentRank constructs a PercentRank or panics if the period is
// invalid.
func MustNewPercentRank(period int) *PercentRank {
	p, err := NewPercentRank(period)
	if err != nil {
		panic(err)
	}
	return p
}

// Process consumes a value x and returns
//
//	(#others < x + ½·#others == x) / (period - 1)
//
// where the others are the period-1 values before x in the window. The
// result is in [0, 1]: 0 when x is below all of them, 1 when above all of
// them, and ties count half. The first value is ready on the period-th
// input. A flat window returns 0.5, and so does a period of 1, which has no
// other values to compare against. NaN inputs are not supported: the result
// is undefined once a NaN has been input.
func (p *PercentRank) Process(v float64) (float64, bool) {
	if !p.window.Push(v) {
		return 0, false
	}
	others := p.window.period - 1
	if others == 0 {
		return 0.5, true
	}
	below := p.window.countBelow(v)
	equal := p.window.countEqual(v) - 1 // v itself is in the window
	return (float64(below) + 0.5*float64(equal)) / float64(others), true
}

// Reset returns the PercentRank to its newly constructed state.
func (p *PercentRank) Reset() {
	p.window.Reset()
}
