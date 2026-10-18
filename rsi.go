package ta

// RSI computes the Relative Strength Index over a fixed period.
type RSI struct {
	prev    float64
	hasPrev bool
	gain    *wilderState
	loss    *wilderState
}

// NewRSI constructs an RSI with the given period.
func NewRSI(period int) (*RSI, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RSI{gain: newWilderState(period), loss: newWilderState(period)}, nil
}

// MustNewRSI constructs an RSI or panics if the period is invalid.
func MustNewRSI(period int) *RSI {
	rsi, err := NewRSI(period)
	if err != nil {
		panic(err)
	}
	return rsi
}

// Process consumes a value and returns the RSI once enough values are available.
func (r *RSI) Process(v float64) (float64, bool) {
	if !r.hasPrev {
		r.prev = v
		r.hasPrev = true
		return 0, false
	}

	diff := v - r.prev
	r.prev = v

	gain := 0.0
	loss := 0.0
	if diff > 0 {
		gain = diff
	} else if diff < 0 {
		loss = -diff
	}

	avgGain, ready := r.gain.Push(gain)
	avgLoss, _ := r.loss.Push(loss)
	if !ready {
		return 0, false
	}
	return rsiFromAvg(avgGain, avgLoss), true
}

func rsiFromAvg(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		if avgGain == 0 {
			return 50
		}
		return 100
	}
	if avgGain == 0 {
		return 0
	}
	rs := avgGain / avgLoss
	return 100 - (100 / (1 + rs))
}

// Reset returns the RSI to its newly constructed state.
func (r *RSI) Reset() {
	r.prev, r.hasPrev = 0, false
	r.gain.Reset()
	r.loss.Reset()
}
