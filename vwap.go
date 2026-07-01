package ta

// VWAP computes the cumulative Volume Weighted Average Price.
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

// Update consumes an OHLCV value and returns the running VWAP.
func (v *VWAP) Process(value OHLCV) (float64, bool) {
	typicalPrice := (value.High + value.Low + value.Close) / 3
	v.cumulativePV += typicalPrice * value.Volume
	v.cumulativeVolume += value.Volume
	if v.cumulativeVolume == 0 {
		return 0, true
	}
	return v.cumulativePV / v.cumulativeVolume, true
}
