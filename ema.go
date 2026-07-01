package ta

// EMA computes the Exponential Moving Average over a fixed period.
type EMA struct {
	state *emaState
}

// emaState maintains a rolling exponential moving average.
type emaState struct {
	period int
	alpha  float64
	count  int
	sum    float64
	value  float64
	ready  bool
}

func newEMAState(period int) *emaState {
	return &emaState{
		period: period,
		alpha:  2 / (float64(period) + 1),
	}
}

// Push adds a value and returns the EMA when ready.
func (e *emaState) Push(v float64) (float64, bool) {
	if !e.ready {
		e.sum += v
		e.count++
		if e.count == e.period {
			e.value = e.sum / float64(e.period)
			e.ready = true
			return e.value, true
		}
		return 0, false
	}

	e.value = (v-e.value)*e.alpha + e.value
	return e.value, true
}

// NewEMA constructs an EMA with the given period.
func NewEMA(period int) (*EMA, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &EMA{state: newEMAState(period)}, nil
}

// MustNewEMA constructs an EMA or panics if the period is invalid.
func MustNewEMA(period int) *EMA {
	ema, err := NewEMA(period)
	if err != nil {
		panic(err)
	}
	return ema
}

// Update consumes a value and returns the EMA once enough values are available.
func (e *EMA) Process(v float64) (float64, bool) {
	return e.state.Push(v)
}
