package ta

import (
	"fmt"
	"testing"
)

// benchProcessor feeds input to p in a loop, b.N values in all.
func benchProcessor[I, O any](b *testing.B, p Processor[I, O], input []I) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Process(input[i%len(input)])
	}
}

var benchPeriods = []int{14, 200}

// orderStatPeriods covers the window sizes the sorted window is used at:
// 20 and 60 bars, a trading day of minutes (390), and a full day (1440).
var orderStatPeriods = []int{20, 60, 390, 1440}

func BenchmarkStoch(b *testing.B) {
	input := randomOHLC(1, 4096)
	for _, k := range benchPeriods {
		b.Run(fmt.Sprintf("k=%d", k), func(b *testing.B) {
			benchProcessor(b, MustNewStoch(k, 3), input)
		})
	}
}

func BenchmarkMax(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range benchPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewMax(p), input) })
	}
}

func BenchmarkMin(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range benchPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewMin(p), input) })
	}
}

func BenchmarkSum(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range benchPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewSum(p), input) })
	}
}

func BenchmarkLag(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range benchPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewLag(p), input) })
	}
}

func BenchmarkTrueRange(b *testing.B) {
	benchProcessor(b, MustNewTrueRange(), randomOHLC(1, 4096))
}

func BenchmarkMedian(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range orderStatPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewMedian(p), input) })
	}
}

func BenchmarkMAD(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range orderStatPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewMAD(p), input) })
	}
}

func BenchmarkPercentRank(b *testing.B) {
	input := randomWalk(1, 4096)
	for _, p := range orderStatPeriods {
		b.Run(fmt.Sprintf("period=%d", p), func(b *testing.B) { benchProcessor(b, MustNewPercentRank(p), input) })
	}
}
