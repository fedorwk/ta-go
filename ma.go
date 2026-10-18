package ta

// MAType selects the moving average an indicator uses for smoothing.
type MAType int

const (
	MATypeSMA MAType = iota // Simple moving average
	MATypeEMA               // Exponential moving average, seeded with a simple average
	MATypeWMA               // Linearly weighted moving average
	MATypeRMA               // Wilder's smoothed moving average (also called SMMA)
)

// String returns the moving average's abbreviation, such as "SMA".
func (t MAType) String() string {
	switch t {
	case MATypeSMA:
		return "SMA"
	case MATypeEMA:
		return "EMA"
	case MATypeWMA:
		return "WMA"
	case MATypeRMA:
		return "RMA"
	}
	return "MAType(invalid)"
}

func (t MAType) valid() bool {
	return t >= MATypeSMA && t <= MATypeRMA
}

// maState is a moving average that can be selected by MAType. Every type
// emits its first value after period inputs.
type maState interface {
	Push(v float64) (float64, bool)
	Reset()
}

var (
	_ maState = (*smaState)(nil)
	_ maState = (*emaState)(nil)
	_ maState = (*wmaState)(nil)
	_ maState = (*wilderState)(nil)
)

// newMAState returns a moving average of the given type. The type must be
// valid.
func newMAState(period int, t MAType) maState {
	switch t {
	case MATypeSMA:
		return newSMAState(period)
	case MATypeEMA:
		return newEMAState(period)
	case MATypeWMA:
		return newWMAState(period)
	case MATypeRMA:
		return newWilderState(period)
	}
	panic("ta: invalid MAType")
}

// MA computes a moving average whose type is chosen at construction, so
// callers can select it at runtime. It is the only public form of Wilder's
// moving average (MATypeRMA).
type MA struct {
	state maState
}

// NewMA constructs a moving average of the given type and period.
func NewMA(period int, maType MAType) (*MA, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	if !maType.valid() {
		return nil, ErrInvalidMAType
	}
	return &MA{state: newMAState(period, maType)}, nil
}

// MustNewMA constructs an MA or panics if the period or type is invalid.
func MustNewMA(period int, maType MAType) *MA {
	ma, err := NewMA(period, maType)
	if err != nil {
		panic(err)
	}
	return ma
}

// Process consumes a value and returns the moving average once period values
// are available.
func (m *MA) Process(v float64) (float64, bool) {
	return m.state.Push(v)
}

// Reset returns the MA to its newly constructed state.
func (m *MA) Reset() {
	m.state.Reset()
}
