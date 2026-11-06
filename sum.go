package ta

// Sum computes the rolling sum of the last period values.
type Sum struct {
	state *smaState
}

// NewSum constructs a Sum with the given period.
func NewSum(period int) (*Sum, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Sum{state: newSMAState(period)}, nil
}

// MustNewSum constructs a Sum or panics if the period is invalid.
func MustNewSum(period int) *Sum {
	s, err := NewSum(period)
	if err != nil {
		panic(err)
	}
	return s
}

// Process consumes a value and returns x[t-period+1] + ... + x[t], the sum of
// the last period values. The first value is ready on the period-th input.
// There are no degenerate inputs: a window of zeros sums to 0. The running sum
// is recomputed exactly from the window every period inputs, so it does not
// drift over long streams.
func (s *Sum) Process(v float64) (float64, bool) {
	if !s.state.add(v) {
		return 0, false
	}
	return s.state.sum, true
}

// Reset returns the Sum to its newly constructed state.
func (s *Sum) Reset() {
	s.state.Reset()
}

// nonNegSumState maintains the rolling sum of the last period non-negative
// terms, such as squared returns or absolute changes.
//
// The underlying smaState recomputes the sum exactly each time its ring
// wraps. Between wraps, each update rounds the sum by about one ulp of its
// current size, so when a large term leaves the window the residue of the
// larger sum can swamp the new one. As in varianceState, the sum is therefore
// also recomputed whenever it falls below m2DropLimit of its peak since the
// last exact computation. That needs the sum to collapse by that factor, so
// it is rare and the cost stays amortized O(1). A window of zeros therefore
// sums to exactly 0.
type nonNegSumState struct {
	state *smaState
	peak  float64 // the largest sum since its last exact computation
}

func newNonNegSumState(period int) *nonNegSumState {
	return &nonNegSumState{state: newSMAState(period)}
}

// add adds a term, which must be ≥ 0, and reports whether the window is full,
// in which case sum returns the sum of the last period terms.
func (s *nonNegSumState) add(v float64) bool {
	if !s.state.add(v) {
		return false
	}
	switch {
	case s.state.idx == 0: // the window just filled, or wrapped and was recomputed
		s.peak = s.state.sum
	case s.state.sum < s.peak*m2DropLimit:
		s.state.recompute()
		s.peak = s.state.sum
	default:
		s.peak = max(s.peak, s.state.sum)
	}
	return true
}

// sum returns the sum of the window, clamped at 0 against rounding.
func (s *nonNegSumState) sum() float64 {
	return max(s.state.sum, 0)
}

// Reset empties the window.
func (s *nonNegSumState) Reset() {
	s.state.Reset()
	s.peak = 0
}
