package ta

// MACD computes the Moving Average Convergence Divergence.
type MACD struct {
	fast          int
	slow          int
	signalPeriod  int
	fastEMA       *emaState
	slowEMA       *emaState
	signalAlpha   float64
	signal        float64
	signalCount   int
	seenFirstMACD bool
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
		fast:         fast,
		slow:         slow,
		signalPeriod: signal,
		fastEMA:      newEMAState(fast),
		slowEMA:      newEMAState(slow),
		signalAlpha:  2 / (float64(signal) + 1),
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

// Update consumes a value and returns a MACDValue once enough values are available.
func (m *MACD) Process(v float64) (MACDValue, bool) {
	fast, fastReady := m.fastEMA.Push(v)
	slow, slowReady := m.slowEMA.Push(v)
	if !fastReady || !slowReady {
		return MACDValue{}, false
	}

	macd := fast - slow
	if !m.seenFirstMACD {
		m.seenFirstMACD = true
		return MACDValue{}, false
	}

	m.signal = (macd-m.signal)*m.signalAlpha + m.signal
	m.signalCount++
	if m.signalCount < m.signalPeriod-1 {
		return MACDValue{}, false
	}

	return MACDValue{
		MACD:   macd,
		Signal: m.signal,
		Hist:   macd - m.signal,
	}, true
}
