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

// Lag delays a series by a fixed number of inputs. Chain it in front of
// another processor to compute over the values before the current one: for
// example, Chain(MustNewLag(1), MustNewMax(20)) is the highest of the 20
// values before the current one.
type Lag struct {
	state *lagState
}

// NewLag constructs a Lag with the given period.
func NewLag(period int) (*Lag, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Lag{state: newLagState(period)}, nil
}

// MustNewLag constructs a Lag or panics if the period is invalid.
func MustNewLag(period int) *Lag {
	l, err := NewLag(period)
	if err != nil {
		panic(err)
	}
	return l
}

// Process consumes a value and returns x[t-period], the value period inputs
// earlier. The first value is ready on input period+1. Values pass through
// unchanged, so there are no degenerate inputs.
func (l *Lag) Process(v float64) (float64, bool) {
	return l.state.Push(v)
}

// Reset returns the Lag to its newly constructed state.
func (l *Lag) Reset() {
	l.state.Reset()
}
