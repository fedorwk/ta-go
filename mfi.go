package ta

// MFI computes the Money Flow Index over a fixed period.
type MFI struct {
	prevTP  float64
	hasPrev bool
	pos     *smaState // rolling sums of positive and negative money flow
	neg     *smaState
}

// NewMFI constructs an MFI with the given period.
func NewMFI(period int) (*MFI, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &MFI{pos: newSMAState(period), neg: newSMAState(period)}, nil
}

// MustNewMFI constructs an MFI or panics if the period is invalid.
func MustNewMFI(period int) *MFI {
	mfi, err := NewMFI(period)
	if err != nil {
		panic(err)
	}
	return mfi
}

// Process consumes an OHLCV value and returns the MFI once enough values are
// available.
//
// Raw money flow is typical price times volume, counted as positive when the
// typical price rose and negative when it fell. Like TA-Lib, the MFI is 0
// when the window's total money flow is below 1.
func (m *MFI) Process(v OHLCV) (float64, bool) {
	tp := (v.High + v.Low + v.Close) / 3
	if !m.hasPrev {
		m.prevTP = tp
		m.hasPrev = true
		return 0, false
	}

	flow := tp * v.Volume
	pos, neg := 0.0, 0.0
	if tp > m.prevTP {
		pos = flow
	} else if tp < m.prevTP {
		neg = flow
	}
	m.prevTP = tp

	_, ready := m.pos.Push(pos)
	m.neg.Push(neg)
	if !ready {
		return 0, false
	}
	total := m.pos.sum + m.neg.sum
	if total < 1 {
		return 0, true
	}
	return 100 * m.pos.sum / total, true
}

// Reset returns the MFI to its newly constructed state.
func (m *MFI) Reset() {
	m.prevTP, m.hasPrev = 0, false
	m.pos.Reset()
	m.neg.Reset()
}
