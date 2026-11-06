package ta

// extremumState maintains the maximum or minimum of the last period values in
// O(1) amortized time per value, using a monotonic deque stored in a ring
// buffer.
type extremumState struct {
	period  int
	keepMax bool
	count   int // values pushed so far; also the index of the next value
	buf     []extremumEntry
	head    int
	size    int
}

type extremumEntry struct {
	idx int
	v   float64
}

func newMaxState(period int) *extremumState {
	return &extremumState{period: period, keepMax: true, buf: make([]extremumEntry, period)}
}

func newMinState(period int) *extremumState {
	return &extremumState{period: period, buf: make([]extremumEntry, period)}
}

// Push adds a value and returns the extremum of the last period values once
// period values have been pushed.
func (e *extremumState) Push(v float64) (float64, bool) {
	// Drop the front entry if it falls out of the window once v is added.
	// This also guarantees room for v in the ring buffer.
	if e.size > 0 && e.buf[e.head].idx <= e.count-e.period {
		e.head = (e.head + 1) % e.period
		e.size--
	}
	// Drop entries v dominates: they can never be the extremum again.
	for e.size > 0 {
		back := e.buf[(e.head+e.size-1)%e.period].v
		if e.keepMax && back > v || !e.keepMax && back < v {
			break
		}
		e.size--
	}
	e.buf[(e.head+e.size)%e.period] = extremumEntry{idx: e.count, v: v}
	e.size++
	e.count++

	if e.count < e.period {
		return 0, false
	}
	return e.buf[e.head].v, true
}

// Reset empties the window.
func (e *extremumState) Reset() {
	e.count, e.head, e.size = 0, 0, 0
}

// Max computes the highest of the last period values.
type Max struct {
	state *extremumState
}

// NewMax constructs a Max with the given period.
func NewMax(period int) (*Max, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Max{state: newMaxState(period)}, nil
}

// MustNewMax constructs a Max or panics if the period is invalid.
func MustNewMax(period int) *Max {
	m, err := NewMax(period)
	if err != nil {
		panic(err)
	}
	return m
}

// Process consumes a value and returns max(x[t-period+1], ..., x[t]), the
// highest of the last period values. The first value is ready on the
// period-th input. A flat window returns its value. NaN inputs are not
// supported: the result is undefined while a NaN is in the window.
func (m *Max) Process(v float64) (float64, bool) {
	return m.state.Push(v)
}

// Reset returns the Max to its newly constructed state.
func (m *Max) Reset() {
	m.state.Reset()
}

// Min computes the lowest of the last period values.
type Min struct {
	state *extremumState
}

// NewMin constructs a Min with the given period.
func NewMin(period int) (*Min, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Min{state: newMinState(period)}, nil
}

// MustNewMin constructs a Min or panics if the period is invalid.
func MustNewMin(period int) *Min {
	m, err := NewMin(period)
	if err != nil {
		panic(err)
	}
	return m
}

// Process consumes a value and returns min(x[t-period+1], ..., x[t]), the
// lowest of the last period values. The first value is ready on the period-th
// input. A flat window returns its value. NaN inputs are not supported: the
// result is undefined while a NaN is in the window.
func (m *Min) Process(v float64) (float64, bool) {
	return m.state.Push(v)
}

// Reset returns the Min to its newly constructed state.
func (m *Min) Reset() {
	m.state.Reset()
}
