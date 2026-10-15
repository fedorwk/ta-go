package ta

// OBV computes On-Balance Volume.
type OBV struct {
	prevClose float64
	hasPrev   bool
	obv       float64
}

// NewOBV constructs OBV.
func NewOBV() *OBV {
	return &OBV{}
}

// MustNewOBV constructs OBV.
func MustNewOBV() *OBV {
	return NewOBV()
}

// Process consumes an OHLCV value and returns the running OBV.
func (o *OBV) Process(v OHLCV) (float64, bool) {
	if !o.hasPrev {
		o.obv = v.Volume
		o.prevClose = v.Close
		o.hasPrev = true
		return o.obv, true
	}

	if v.Close > o.prevClose {
		o.obv += v.Volume
	} else if v.Close < o.prevClose {
		o.obv -= v.Volume
	}

	o.prevClose = v.Close
	return o.obv, true
}
