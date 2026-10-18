package ta

import (
	"fmt"
	"testing"
)

func BenchmarkStoch(b *testing.B) {
	input := randomOHLC(1, 4096)
	for _, k := range []int{14, 200} {
		b.Run(fmt.Sprintf("k=%d", k), func(b *testing.B) {
			st := MustNewStoch(k, 3)
			for i := 0; i < b.N; i++ {
				st.Process(input[i%len(input)])
			}
		})
	}
}
