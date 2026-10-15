package ta

// MACD computes the Moving Average Convergence Divergence.
//
// Values match TA-Lib: both EMAs and the signal line are seeded with simple
// averages, and the fast EMA is seeded with the mean of the fast-period values
// ending where the slow EMA becomes ready, so both lines start on the same
// input. The first value is emitted after slow+signal-1 inputs.
type MACD struct {
	fastSeed  *smaState // rolling mean used to seed the fast EMA; nil once seeded
	fastAlpha float64
	fast      float64
	slowEMA   *emaState
	signalEMA *emaState
}

// MACDValue contains the MACD line, signal line, and histogram.
type MACDValue struct {
	MACD   float64
	Signal float64
	Hist   float64
}

// NewMACD constructs a MACD with the given fast, slow, and signal periods.
func NewMACD(fast, slow, signal int) (*MACD, error) {
	if fast <= 0 || slow <= 0 || signal <= 0 || fast >= slow {
		return nil, ErrInvalidMACDPeriods
	}
	return &MACD{
		fastSeed:  newSMAState(fast),
		fastAlpha: 2 / (float64(fast) + 1),
		slowEMA:   newEMAState(slow),
		signalEMA: newEMAState(signal),
	}, nil
}

// MustNewMACD constructs a MACD or panics if the periods are invalid.
func MustNewMACD(fast, slow, signal int) *MACD {
	macd, err := NewMACD(fast, slow, signal)
	if err != nil {
		panic(err)
	}
	return macd
}

// Process consumes a value and returns a MACDValue once enough values are available.
func (m *MACD) Process(v float64) (MACDValue, bool) {
	slow, slowReady := m.slowEMA.Push(v)
	if m.fastSeed != nil {
		seed, _ := m.fastSeed.Push(v)
		if !slowReady {
			return MACDValue{}, false
		}
		m.fast = seed
		m.fastSeed = nil
	} else {
		m.fast = (v-m.fast)*m.fastAlpha + m.fast
	}

	macd := m.fast - slow
	signal, ok := m.signalEMA.Push(macd)
	if !ok {
		return MACDValue{}, false
	}
	return MACDValue{
		MACD:   macd,
		Signal: signal,
		Hist:   macd - signal,
	}, true
}
