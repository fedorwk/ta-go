package ta

// Bollinger computes Bollinger Bands over a fixed period.
type Bollinger struct {
	stddev float64
	state  *varianceState
	ma     maState // middle band; nil for SMA, which uses state's mean
}

// Bands contains the upper, middle, and lower bands of a channel indicator,
// with the latest price's position in the channel and the channel's width.
type Bands struct {
	Upper  float64
	Middle float64
	Lower  float64
	// PercentB is (price-Lower)/(Upper-Lower): 0 at the lower band, 1 at the
	// upper band, outside [0, 1] beyond them, and 0.5 when the bands coincide.
	PercentB float64
	// Bandwidth is (Upper-Lower)/Middle, or 0 when Middle is 0.
	Bandwidth float64
}

// newBands returns Bands with PercentB and Bandwidth computed for price.
func newBands(upper, middle, lower, price float64) Bands {
	b := Bands{Upper: upper, Middle: middle, Lower: lower, PercentB: 0.5}
	if width := upper - lower; width != 0 {
		b.PercentB = (price - lower) / width
	}
	if middle != 0 {
		b.Bandwidth = (upper - lower) / middle
	}
	return b
}

// NewBollinger constructs Bollinger Bands with the given period and stddev
// multiplier, around a simple moving average.
func NewBollinger(period int, stddev float64) (*Bollinger, error) {
	return NewBollingerMA(period, stddev, MATypeSMA)
}

// MustNewBollinger constructs Bollinger Bands or panics if parameters are invalid.
func MustNewBollinger(period int, stddev float64) *Bollinger {
	bb, err := NewBollinger(period, stddev)
	if err != nil {
		panic(err)
	}
	return bb
}

// NewBollingerMA constructs Bollinger Bands whose middle band is the given
// moving average. As in TA-Lib's BBANDS, the band width is still stddev
// times the standard deviation around the simple mean.
func NewBollingerMA(period int, stddev float64, maType MAType) (*Bollinger, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	if stddev <= 0 {
		return nil, ErrInvalidStdDev
	}
	if !maType.valid() {
		return nil, ErrInvalidMAType
	}
	bb := &Bollinger{
		stddev: stddev,
		state:  newVarianceState(period),
	}
	if maType != MATypeSMA {
		bb.ma = newMAState(period, maType)
	}
	return bb, nil
}

// MustNewBollingerMA constructs Bollinger Bands or panics if parameters are invalid.
func MustNewBollingerMA(period int, stddev float64, maType MAType) *Bollinger {
	bb, err := NewBollingerMA(period, stddev, maType)
	if err != nil {
		panic(err)
	}
	return bb
}

// Process consumes a value and returns Bollinger Bands once enough values are
// available. PercentB places v within the bands.
//
// Bands use the population standard deviation, as TA-Lib's BBANDS does. It is
// maintained with Welford's algorithm, so it stays accurate when the values
// are large relative to their spread.
func (b *Bollinger) Process(v float64) (Bands, bool) {
	mean := 0.0
	if b.ma != nil {
		mean, _ = b.ma.Push(v)
	}
	if !b.state.Push(v) {
		return Bands{}, false
	}
	if b.ma == nil {
		mean = b.state.mean
	}
	width := b.stddev * b.state.stdDev()
	return newBands(mean+width, mean, mean-width, v), true
}

// Reset returns the bands to their newly constructed state.
func (b *Bollinger) Reset() {
	b.state.Reset()
	if b.ma != nil {
		b.ma.Reset()
	}
}
