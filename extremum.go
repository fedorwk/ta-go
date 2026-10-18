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
