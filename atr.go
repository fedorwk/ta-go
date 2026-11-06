package ta

// ATR computes the Average True Range over a fixed period.
type ATR struct {
	tr    TrueRange
	state *wilderState
}

// NewATR constructs an ATR with the given period.
func NewATR(period int) (*ATR, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &ATR{state: newWilderState(period)}, nil
}

// MustNewATR constructs an ATR or panics if parameters are invalid.
func MustNewATR(period int) *ATR {
	atr, err := NewATR(period)
	if err != nil {
		panic(err)
	}
	return atr
}

// Process consumes an OHLC value and returns the ATR once enough values are available.
func (a *ATR) Process(v OHLC) (float64, bool) {
	tr, ok := a.tr.Process(v)
	if !ok {
		return 0, false
	}
	return a.state.Push(tr)
}

// Reset returns the ATR to its newly constructed state.
func (a *ATR) Reset() {
	a.tr.Reset()
	a.state.Reset()
}
