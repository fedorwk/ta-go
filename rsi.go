package ta

// RSI computes the Relative Strength Index over a fixed period.
type RSI struct {
	period  int
	prev    float64
	hasPrev bool
	gains   float64
	losses  float64
	count   int
	avgGain float64
	avgLoss float64
	ready   bool
}

// NewRSI constructs an RSI with the given period.
func NewRSI(period int) (*RSI, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &RSI{period: period}, nil
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

	if !r.ready {
		r.gains += gain
		r.losses += loss
		r.count++
		if r.count == r.period {
			r.avgGain = r.gains / float64(r.period)
			r.avgLoss = r.losses / float64(r.period)
			r.ready = true
			return rsiFromAvg(r.avgGain, r.avgLoss), true
		}
		return 0, false
	}

	r.avgGain = (r.avgGain*float64(r.period-1) + gain) / float64(r.period)
	r.avgLoss = (r.avgLoss*float64(r.period-1) + loss) / float64(r.period)
	return rsiFromAvg(r.avgGain, r.avgLoss), true
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
