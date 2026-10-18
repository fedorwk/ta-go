package ta

// Stoch computes the Stochastic Oscillator.
type Stoch struct {
	highest *extremumState
	lowest  *extremumState
	slowK   maState // smooths fast %K into slow %K; nil when unsmoothed
	dState  maState
}

// StochValue contains the %K and %D values.
type StochValue struct {
	K float64
	D float64
}

// NewStoch constructs a fast Stochastic Oscillator with the given %K and %D
// periods: %K is unsmoothed and %D is its simple moving average.
func NewStoch(kPeriod, dPeriod int) (*Stoch, error) {
	return NewStochFull(kPeriod, 1, dPeriod, MATypeSMA)
}

// MustNewStoch constructs a Stochastic Oscillator or panics if parameters are invalid.
func MustNewStoch(kPeriod, dPeriod int) *Stoch {
	st, err := NewStoch(kPeriod, dPeriod)
	if err != nil {
		panic(err)
	}
	return st
}

// NewStochFull constructs a full Stochastic Oscillator: the fast %K over
// kPeriod is smoothed over slowKPeriod into the reported %K, and %D smooths
// that over dPeriod, both with the given moving average. A slowKPeriod of 1
// gives the fast oscillator; NewStochFull(14, 3, 3, MATypeSMA) is the
// common slow oscillator. Values match TA-Lib's STOCH with both MA types set
// to maType.
func NewStochFull(kPeriod, slowKPeriod, dPeriod int, maType MAType) (*Stoch, error) {
	if kPeriod <= 0 || slowKPeriod <= 0 || dPeriod <= 0 {
		return nil, ErrInvalidPeriod
	}
	if !maType.valid() {
		return nil, ErrInvalidMAType
	}
	st := &Stoch{
		highest: newMaxState(kPeriod),
		lowest:  newMinState(kPeriod),
		dState:  newMAState(dPeriod, maType),
	}
	if slowKPeriod > 1 {
		st.slowK = newMAState(slowKPeriod, maType)
	}
	return st, nil
}

// MustNewStochFull constructs a full Stochastic Oscillator or panics if
// parameters are invalid.
func MustNewStochFull(kPeriod, slowKPeriod, dPeriod int, maType MAType) *Stoch {
	st, err := NewStochFull(kPeriod, slowKPeriod, dPeriod, maType)
	if err != nil {
		panic(err)
	}
	return st
}

// Process consumes an OHLC value and returns %K/%D once enough values are
// available: after kPeriod+slowKPeriod+dPeriod-2 inputs.
func (s *Stoch) Process(v OHLC) (StochValue, bool) {
	highest, ready := s.highest.Push(v.High)
	lowest, _ := s.lowest.Push(v.Low)
	if !ready {
		return StochValue{}, false
	}

	k := 0.0
	denom := highest - lowest
	if denom != 0 {
		k = 100 * (v.Close - lowest) / denom
	}
	if s.slowK != nil {
		if k, ready = s.slowK.Push(k); !ready {
			return StochValue{}, false
		}
	}

	d, ready := s.dState.Push(k)
	if !ready {
		return StochValue{}, false
	}
	return StochValue{K: k, D: d}, true
}

// Reset returns the oscillator to its newly constructed state.
func (s *Stoch) Reset() {
	s.highest.Reset()
	s.lowest.Reset()
	if s.slowK != nil {
		s.slowK.Reset()
	}
	s.dState.Reset()
}
