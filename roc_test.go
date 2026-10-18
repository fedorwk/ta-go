package ta

import (
	"testing"

	talib "github.com/markcheno/go-talib"
)

func TestROC(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{1, 2, 10, 50} {
		got := apply(input, MustNewROC(period).Process)
		assertAlignedApprox(t, got, talib.Roc(input, period), period, 1e-9)
	}
}

func TestMom(t *testing.T) {
	input := randomWalk(1, 300)
	for _, period := range []int{1, 2, 10, 50} {
		got := apply(input, MustNewMom(period).Process)
		assertAlignedApprox(t, got, talib.Mom(input, period), period, 1e-9)
	}
}
