package ta

// Donchian computes Donchian Channels: the highest high and lowest low over a
// fixed period, with their midpoint.
type Donchian struct {
	highest *extremumState
	lowest  *extremumState
}

// NewDonchian constructs Donchian Channels with the given period.
func NewDonchian(period int) (*Donchian, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &Donchian{highest: newMaxState(period), lowest: newMinState(period)}, nil
}

// MustNewDonchian constructs Donchian Channels or panics if the period is invalid.
func MustNewDonchian(period int) *Donchian {
	dc, err := NewDonchian(period)
	if err != nil {
		panic(err)
	}
	return dc
}

// Process consumes an OHLC value and returns the channel once enough values
// are available. PercentB places the close within the channel.
func (d *Donchian) Process(v OHLC) (Bands, bool) {
	upper, ready := d.highest.Push(v.High)
	lower, _ := d.lowest.Push(v.Low)
	if !ready {
		return Bands{}, false
	}
	return newBands(upper, (upper+lower)/2, lower, v.Close), true
}

// Reset returns the channel to its newly constructed state.
func (d *Donchian) Reset() {
	d.highest.Reset()
	d.lowest.Reset()
}
