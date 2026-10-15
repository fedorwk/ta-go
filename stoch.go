package ta

// Stoch computes the Stochastic Oscillator.
type Stoch struct {
	kPeriod int
	dPeriod int

	window []OHLC
	count  int
	idx    int
	dState *smaState
}

// StochValue contains the %K and %D values.
type StochValue struct {
	K float64
	D float64
}

// NewStoch constructs a Stochastic Oscillator with the given %K and %D periods.
func NewStoch(kPeriod, dPeriod int) (*Stoch, error) {
	if kPeriod <= 0 || dPeriod <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Stoch{
		kPeriod: kPeriod,
		dPeriod: dPeriod,
		window:  make([]OHLC, kPeriod),
		dState:  newSMAState(dPeriod),
	}, nil
}

// MustNewStoch constructs a Stochastic Oscillator or panics if parameters are invalid.
func MustNewStoch(kPeriod, dPeriod int) *Stoch {
	st, err := NewStoch(kPeriod, dPeriod)
	if err != nil {
		panic(err)
	}
	return st
}

// Process consumes an OHLC value and returns %K/%D once enough values are available.
func (s *Stoch) Process(v OHLC) (StochValue, bool) {
	if s.count < s.kPeriod {
		s.window[s.count] = v
		s.count++
		if s.count < s.kPeriod {
			return StochValue{}, false
		}
	} else {
		s.window[s.idx] = v
		s.idx++
		if s.idx == s.kPeriod {
			s.idx = 0
		}
	}

	highest := s.window[0].High
	lowest := s.window[0].Low
	for i := 1; i < s.kPeriod; i++ {
		if s.window[i].High > highest {
			highest = s.window[i].High
		}
		if s.window[i].Low < lowest {
			lowest = s.window[i].Low
		}
	}

	k := 0.0
	denom := highest - lowest
	if denom != 0 {
		k = 100 * (v.Close - lowest) / denom
	}

	d, ready := s.dState.Push(k)
	if !ready {
		return StochValue{}, false
	}
	return StochValue{K: k, D: d}, true
}
