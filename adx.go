package ta

import "math"

// ADX computes the Average Directional Index along with the +DI and -DI
// directional indicators.
//
// Values match TA-Lib: directional movement and true range are summed over
// the first period-1 bars and then smoothed with Wilder's method, so +DI and
// -DI are defined from input period, and ADX, Wilder's average of DX, is
// emitted from input 2*period-1. Every output carries all three values.
type ADX struct {
	period   float64
	seedLeft int // bars still to sum before smoothing starts

	prev    OHLC
	hasPrev bool
	plusDM  float64
	minusDM float64
	tr      float64

	adxState *wilderState
	adx      float64
}

// ADXValue contains the ADX and the directional indicators it is built from.
type ADXValue struct {
	ADX     float64
	PlusDI  float64
	MinusDI float64
}

// NewADX constructs an ADX with the given period.
func NewADX(period int) (*ADX, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &ADX{
		period:   float64(period),
		seedLeft: period - 1,
		adxState: newWilderState(period),
	}, nil
}

// MustNewADX constructs an ADX or panics if the period is invalid.
func MustNewADX(period int) *ADX {
	adx, err := NewADX(period)
	if err != nil {
		panic(err)
	}
	return adx
}

// Process consumes an OHLC value and returns the ADX and directional
// indicators once enough values are available.
func (a *ADX) Process(v OHLC) (ADXValue, bool) {
	if !a.hasPrev {
		a.prev = v
		a.hasPrev = true
		return ADXValue{}, false
	}

	upMove := v.High - a.prev.High
	downMove := a.prev.Low - v.Low
	plusDM, minusDM := 0.0, 0.0
	if downMove > 0 && upMove < downMove {
		minusDM = downMove
	} else if upMove > 0 && upMove > downMove {
		plusDM = upMove
	}
	tr := trueRange(v, a.prev.Close)
	a.prev = v

	if a.seedLeft > 0 {
		a.seedLeft--
		a.plusDM += plusDM
		a.minusDM += minusDM
		a.tr += tr
		return ADXValue{}, false
	}

	a.plusDM = a.plusDM - a.plusDM/a.period + plusDM
	a.minusDM = a.minusDM - a.minusDM/a.period + minusDM
	a.tr = a.tr - a.tr/a.period + tr

	var plusDI, minusDI, dx float64
	hasDX := false
	if !isNearZero(a.tr) {
		plusDI = 100 * a.plusDM / a.tr
		minusDI = 100 * a.minusDM / a.tr
		if sum := plusDI + minusDI; !isNearZero(sum) {
			dx = 100 * math.Abs(plusDI-minusDI) / sum
			hasDX = true
		}
	}

	// Like TA-Lib, a bar without a defined DX counts as zero while ADX is
	// seeded and leaves ADX unchanged afterwards.
	ready := a.adxState.ready
	if hasDX || !ready {
		a.adx, ready = a.adxState.Push(dx)
	}
	if !ready {
		return ADXValue{}, false
	}
	return ADXValue{ADX: a.adx, PlusDI: plusDI, MinusDI: minusDI}, true
}

// isNearZero reports whether x is within TA-Lib's zero tolerance.
func isNearZero(x float64) bool {
	return math.Abs(x) < 1e-14
}

// Reset returns the ADX to its newly constructed state.
func (a *ADX) Reset() {
	a.seedLeft = int(a.period) - 1
	a.prev, a.hasPrev = OHLC{}, false
	a.plusDM, a.minusDM, a.tr = 0, 0, 0
	a.adxState.Reset()
	a.adx = 0
}
