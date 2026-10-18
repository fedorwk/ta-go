package ta

// wilderState maintains Wilder's smoothed moving average (also called RMA or
// SMMA), as used by RSI, ATR, and ADX. It is seeded with the simple average of
// the first period values; each later value v updates it to
// (prev*(period-1) + v) / period.
type wilderState struct {
	period int
	count  int
	sum    float64
	value  float64
	ready  bool
}

func newWilderState(period int) *wilderState {
	return &wilderState{period: period}
}

// Push adds a value and returns the smoothed average when ready.
func (w *wilderState) Push(v float64) (float64, bool) {
	if !w.ready {
		w.sum += v
		w.count++
		if w.count == w.period {
			w.value = w.sum / float64(w.period)
			w.ready = true
			return w.value, true
		}
		return 0, false
	}

	w.value = (w.value*float64(w.period-1) + v) / float64(w.period)
	return w.value, true
}

// Reset discards all values pushed so far.
func (w *wilderState) Reset() {
	w.count, w.sum, w.value, w.ready = 0, 0, 0, false
}
