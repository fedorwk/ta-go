package ta

// VWAP computes the cumulative Volume Weighted Average Price since
// construction or the last Reset.
type VWAP struct {
	cumulativePV     float64
	cumulativeVolume float64
}

// NewVWAP constructs VWAP.
func NewVWAP() *VWAP {
	return &VWAP{}
}

// MustNewVWAP constructs VWAP.
func MustNewVWAP() *VWAP {
	return NewVWAP()
}

// Process consumes an OHLCV value and returns the running VWAP.
func (v *VWAP) Process(value OHLCV) (float64, bool) {
	typicalPrice := (value.High + value.Low + value.Close) / 3
	v.cumulativePV += typicalPrice * value.Volume
	v.cumulativeVolume += value.Volume
	if v.cumulativeVolume == 0 {
		return 0, true
	}
	return v.cumulativePV / v.cumulativeVolume, true
}

// Reset clears the running totals, starting a new session: the next value
// is the first of the new VWAP. Call it at each session boundary for a
// session (e.g. daily) VWAP.
func (v *VWAP) Reset() {
	v.cumulativePV, v.cumulativeVolume = 0, 0
}
