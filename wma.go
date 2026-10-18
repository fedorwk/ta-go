package ta

// WMA computes the Weighted Moving Average over a fixed period.
type WMA struct {
	state *wmaState
}

type wmaState struct {
	period   int
	denom    float64
	window   []float64
	count    int
	idx      int
	sum      float64
	weighted float64
}

func newWMAState(period int) *wmaState {
	return &wmaState{
		period: period,
		denom:  float64(period*(period+1)) / 2,
		window: make([]float64, period),
	}
}

// Push adds a value and returns the WMA when ready.
func (w *wmaState) Push(v float64) (float64, bool) {
	if w.count < w.period {
		w.count++
		w.window[w.count-1] = v
		w.sum += v
		w.weighted += float64(w.count) * v
		if w.count == w.period {
			return w.weighted / w.denom, true
		}
		return 0, false
	}

	oldest := w.window[w.idx]
	w.window[w.idx] = v
	w.idx++
	if w.idx == w.period {
		w.idx = 0
	}

	w.weighted = w.weighted - w.sum + float64(w.period)*v
	w.sum = w.sum - oldest + v
	return w.weighted / w.denom, true
}

// Reset empties the window.
func (w *wmaState) Reset() {
	w.count, w.idx, w.sum, w.weighted = 0, 0, 0, 0
}

// NewWMA constructs a WMA with the given period.
func NewWMA(period int) (*WMA, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &WMA{state: newWMAState(period)}, nil
}

// MustNewWMA constructs a WMA or panics if the period is invalid.
func MustNewWMA(period int) *WMA {
	wma, err := NewWMA(period)
	if err != nil {
		panic(err)
	}
	return wma
}

// Process consumes a value and returns the WMA once enough values are available.
func (w *WMA) Process(v float64) (float64, bool) {
	return w.state.Push(v)
}

// Reset returns the WMA to its newly constructed state.
func (w *WMA) Reset() {
	w.state.Reset()
}
