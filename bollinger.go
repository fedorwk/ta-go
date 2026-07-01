package ta

import "math"

// Bollinger computes Bollinger Bands over a fixed period.
type Bollinger struct {
	period int
	stddev float64

	window []float64
	sum    float64
	sumsq  float64
	count  int
	idx    int
}

// Bands contains the upper, middle, and lower Bollinger Bands.
type Bands struct {
	Upper  float64
	Middle float64
	Lower  float64
}

// NewBollinger constructs Bollinger Bands with the given period and stddev multiplier.
func NewBollinger(period int, stddev float64) (*Bollinger, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	if stddev <= 0 {
		return nil, ErrInvalidStdDev
	}
	return &Bollinger{
		period: period,
		stddev: stddev,
		window: make([]float64, period),
	}, nil
}

// MustNewBollinger constructs Bollinger Bands or panics if parameters are invalid.
func MustNewBollinger(period int, stddev float64) *Bollinger {
	bb, err := NewBollinger(period, stddev)
	if err != nil {
		panic(err)
	}
	return bb
}

// Update consumes a value and returns Bollinger Bands once enough values are available.
func (b *Bollinger) Process(v float64) (Bands, bool) {
	if b.count < b.period {
		b.window[b.count] = v
		b.sum += v
		b.sumsq += v * v
		b.count++
		if b.count == b.period {
			return bollingerBands(b.sum, b.sumsq, b.period, b.stddev), true
		}
		return Bands{}, false
	}

	old := b.window[b.idx]
	b.sum -= old
	b.sumsq -= old * old
	b.window[b.idx] = v
	b.sum += v
	b.sumsq += v * v
	b.idx++
	if b.idx == b.period {
		b.idx = 0
	}

	return bollingerBands(b.sum, b.sumsq, b.period, b.stddev), true
}

func bollingerBands(sum, sumsq float64, period int, stddev float64) Bands {
	mean := sum / float64(period)
	variance := (sumsq / float64(period)) - (mean * mean)
	if variance < 0 {
		variance = 0
	}
	sd := math.Sqrt(variance)
	return Bands{
		Upper:  mean + stddev*sd,
		Middle: mean,
		Lower:  mean - stddev*sd,
	}
}
