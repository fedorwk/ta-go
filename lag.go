package ta

// lagState returns the value pushed period values ago.
type lagState struct {
	period int
	window []float64
	count  int
	idx    int
}

func newLagState(period int) *lagState {
	return &lagState{period: period, window: make([]float64, period)}
}

// Push adds a value and returns the value pushed period values earlier, once
// one exists.
func (l *lagState) Push(v float64) (float64, bool) {
	old := l.window[l.idx]
	l.window[l.idx] = v
	l.idx++
	if l.idx == l.period {
		l.idx = 0
	}
	if l.count < l.period {
		l.count++
		return 0, false
	}
	return old, true
}

// Reset empties the window.
func (l *lagState) Reset() {
	l.count, l.idx = 0, 0
}
