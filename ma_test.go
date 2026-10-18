package ta

import (
	"errors"
	"testing"

	talib "github.com/markcheno/go-talib"
)

var allMATypes = []MAType{MATypeSMA, MATypeEMA, MATypeWMA, MATypeRMA}

// talibMA returns go-talib's moving average of the given type, zero-padded
// during warm-up. TA-Lib has no Wilder MA type, so RMA uses rmaReference.
func talibMA(input []float64, period int, t MAType) []float64 {
	switch t {
	case MATypeSMA:
		return talib.Ma(input, period, talib.SMA)
	case MATypeEMA:
		return talib.Ma(input, period, talib.EMA)
	case MATypeWMA:
		return talib.Ma(input, period, talib.WMA)
	case MATypeRMA:
		return rmaReference(input, period)
	}
	panic("invalid MAType")
}

// rmaReference computes Wilder's moving average, seeded with the simple
// average of the first period values, zero-padded during warm-up.
func rmaReference(input []float64, period int) []float64 {
	out := make([]float64, len(input))
	if len(input) < period {
		return out
	}
	sum := 0.0
	for _, v := range input[:period] {
		sum += v
	}
	prev := sum / float64(period)
	out[period-1] = prev
	for i := period; i < len(input); i++ {
		prev = (prev*float64(period-1) + input[i]) / float64(period)
		out[i] = prev
	}
	return out
}

func TestMAState(t *testing.T) {
	input := randomWalk(1, 300)
	for _, mt := range allMATypes {
		for _, period := range []int{1, 2, 5, 20} {
			got := apply(input, newMAState(period, mt).Push)
			assertAlignedApprox(t, got, talibMA(input, period, mt), period-1, 1e-9)
		}
	}
}

func TestMA(t *testing.T) {
	input := randomWalk(1, 300)
	for _, mt := range allMATypes {
		for _, period := range []int{1, 2, 5, 20} {
			got := apply(input, MustNewMA(period, mt).Process)
			assertAlignedApprox(t, got, talibMA(input, period, mt), period-1, 1e-9)
		}
	}
}

func TestMAInvalid(t *testing.T) {
	for _, tc := range []struct {
		period int
		mt     MAType
		want   error
	}{
		{0, MATypeSMA, ErrInvalidPeriod},
		{-1, MATypeRMA, ErrInvalidPeriod},
		{14, MAType(-1), ErrInvalidMAType},
		{14, MAType(4), ErrInvalidMAType},
	} {
		if _, err := NewMA(tc.period, tc.mt); !errors.Is(err, tc.want) {
			t.Errorf("NewMA(%d, %d): got %v, want %v", tc.period, tc.mt, err, tc.want)
		}
	}
}

func TestMATypeString(t *testing.T) {
	for _, tc := range []struct {
		t    MAType
		want string
	}{
		{MATypeSMA, "SMA"}, {MATypeEMA, "EMA"}, {MATypeWMA, "WMA"}, {MATypeRMA, "RMA"},
		{MAType(-1), "MAType(invalid)"}, {MAType(4), "MAType(invalid)"},
	} {
		if got := tc.t.String(); got != tc.want {
			t.Errorf("MAType(%d).String() = %q, want %q", int(tc.t), got, tc.want)
		}
	}
}
