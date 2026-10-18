package ta

// StochRSI computes the Stochastic RSI: the fast Stochastic Oscillator
// applied to RSI values.
type StochRSI struct {
	rsi   *RSI
	stoch *Stoch
}

// NewStochRSI constructs a Stochastic RSI with the given RSI period and
// %K and %D periods.
func NewStochRSI(rsiPeriod, kPeriod, dPeriod int) (*StochRSI, error) {
	rsi, err := NewRSI(rsiPeriod)
	if err != nil {
		return nil, err
	}
	stoch, err := NewStoch(kPeriod, dPeriod)
	if err != nil {
		return nil, err
	}
	return &StochRSI{rsi: rsi, stoch: stoch}, nil
}

// MustNewStochRSI constructs a Stochastic RSI or panics if parameters are invalid.
func MustNewStochRSI(rsiPeriod, kPeriod, dPeriod int) *StochRSI {
	s, err := NewStochRSI(rsiPeriod, kPeriod, dPeriod)
	if err != nil {
		panic(err)
	}
	return s
}

// Process consumes a value and returns %K/%D once enough values are
// available: after rsiPeriod+kPeriod+dPeriod-2 inputs.
func (s *StochRSI) Process(v float64) (StochValue, bool) {
	r, ok := s.rsi.Process(v)
	if !ok {
		return StochValue{}, false
	}
	return s.stoch.Process(OHLC{Open: r, High: r, Low: r, Close: r})
}

// Reset returns the Stochastic RSI to its newly constructed state.
func (s *StochRSI) Reset() {
	s.rsi.Reset()
	s.stoch.Reset()
}
