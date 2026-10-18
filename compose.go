package ta

// Chain feeds the outputs of first into second, producing a Processor that
// is ready once both are. For example, Chain(NewRSI(14), NewSMA(5)) yields a
// 5-period average of RSI, and Chain(HLC3, NewSMA(20)) averages typical price.
//
// The returned Processor also has a Reset method, which resets each stage
// that has one; reach it with p.(interface{ Reset() }).
func Chain[A, B, C any](first Processor[A, B], second Processor[B, C]) Processor[A, C] {
	return &chain[A, B, C]{first: first, second: second}
}

type chain[A, B, C any] struct {
	first  Processor[A, B]
	second Processor[B, C]
}

func (c *chain[A, B, C]) Process(v A) (C, bool) {
	b, ok := c.first.Process(v)
	if !ok {
		var zero C
		return zero, false
	}
	return c.second.Process(b)
}

func (c *chain[A, B, C]) Reset() {
	if r, ok := c.first.(interface{ Reset() }); ok {
		r.Reset()
	}
	if r, ok := c.second.(interface{ Reset() }); ok {
		r.Reset()
	}
}

// Apply feeds every input to p and returns the outputs that were ready.
func Apply[I, O any](p Processor[I, O], input []I) []O {
	var out []O
	for _, v := range input {
		if o, ok := p.Process(v); ok {
			out = append(out, o)
		}
	}
	return out
}

// Source selects a price from an OHLC candle. Each Source is a
// Processor[OHLC, float64] that is always ready, so it can be chained in
// front of an indicator that takes plain values.
type Source int

const (
	Open  Source = iota // Open price
	High                // High price
	Low                 // Low price
	Close               // Close price
	HL2                 // (High+Low)/2, the median price
	HLC3                // (High+Low+Close)/3, the typical price
	OHLC4               // (Open+High+Low+Close)/4, the average price
)

// Process returns the selected price of v.
func (s Source) Process(v OHLC) (float64, bool) {
	switch s {
	case Open:
		return v.Open, true
	case High:
		return v.High, true
	case Low:
		return v.Low, true
	case Close:
		return v.Close, true
	case HL2:
		return (v.High + v.Low) / 2, true
	case HLC3:
		return (v.High + v.Low + v.Close) / 3, true
	case OHLC4:
		return (v.Open + v.High + v.Low + v.Close) / 4, true
	}
	panic("ta: invalid Source")
}
