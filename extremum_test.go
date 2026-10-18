package ta

import (
	"math"
	"math/rand"
	"testing"
)

func TestExtremumState(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	inputs := map[string][]float64{
		"random":     randomWalk(1, 500),
		"increasing": make([]float64, 100),
		"decreasing": make([]float64, 100),
		"ties":       make([]float64, 300),
	}
	for i := range inputs["increasing"] {
		inputs["increasing"][i] = float64(i)
		inputs["decreasing"][i] = float64(-i)
	}
	for i := range inputs["ties"] {
		inputs["ties"][i] = float64(r.Intn(3))
	}

	for name, input := range inputs {
		for _, period := range []int{1, 2, 3, 14, 50} {
			maxS, minS := newMaxState(period), newMinState(period)
			for i, v := range input {
				gotMax, okMax := maxS.Push(v)
				gotMin, okMin := minS.Push(v)
				if want := i >= period-1; okMax != want || okMin != want {
					t.Fatalf("%s period %d input %d: ready = %v/%v, want %v", name, period, i, okMax, okMin, want)
				}
				if !okMax {
					continue
				}
				wantMax, wantMin := math.Inf(-1), math.Inf(1)
				for _, w := range input[i-period+1 : i+1] {
					wantMax = math.Max(wantMax, w)
					wantMin = math.Min(wantMin, w)
				}
				if gotMax != wantMax || gotMin != wantMin {
					t.Fatalf("%s period %d input %d: got max %v min %v, want %v %v", name, period, i, gotMax, gotMin, wantMax, wantMin)
				}
			}
		}
	}
}
