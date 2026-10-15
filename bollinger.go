package ta

import "math"

// Bollinger computes Bollinger Bands over a fixed period.
type Bollinger struct {
	period int
	stddev float64

	window []float64
	mean   float64
	m2     float64 // sum of squared deviations from mean over the window
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

// Process consumes a value and returns Bollinger Bands once enough values are available.
//
// The mean and sum of squared deviations are maintained with a sliding-window
// form of Welford's algorithm, which stays accurate when the values are large
// relative to their spread. They are recomputed exactly from the window each
// time it wraps, so rounding error cannot accumulate over long streams.
func (b *Bollinger) Process(v float64) (Bands, bool) {
	if b.count < b.period {
		b.window[b.count] = v
		b.count++
		delta := v - b.mean
		b.mean += delta / float64(b.count)
		b.m2 += delta * (v - b.mean)
		if b.count < b.period {
			return Bands{}, false
		}
		return b.bands(), true
	}

	old := b.window[b.idx]
	b.window[b.idx] = v
	b.idx++
	if b.idx == b.period {
		b.idx = 0
		b.recompute()
		return b.bands(), true
	}

	oldMean := b.mean
	b.mean += (v - old) / float64(b.period)
	b.m2 += (v - old) * (v - b.mean + old - oldMean)
	if b.m2 < 0 {
		b.m2 = 0
	}

	return b.bands(), true
}

// recompute sets mean and m2 from the window with a two-pass computation.
func (b *Bollinger) recompute() {
	sum := 0.0
	for _, x := range b.window {
		sum += x
	}
	b.mean = sum / float64(b.period)
	b.m2 = 0
	for _, x := range b.window {
		d := x - b.mean
		b.m2 += d * d
	}
}

// bands returns the bands for the current window using population variance,
// as TA-Lib's BBANDS does.
func (b *Bollinger) bands() Bands {
	sd := math.Sqrt(b.m2 / float64(b.period))
	return Bands{
		Upper:  b.mean + b.stddev*sd,
		Middle: b.mean,
		Lower:  b.mean - b.stddev*sd,
	}
}
