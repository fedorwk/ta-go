package ta

import "math"

// VWAPPeriod computes the Volume Weighted Average Price over the last period
// candles. For a VWAP anchored at a point in time, such as a session open,
// use VWAP and call Reset at each anchor.
//
// The two running sums are recomputed exactly from the window each time it
// wraps. Between wraps, each update rounds a sum by about one ulp of its
// current size, so when a high-volume candle leaves the window the residue of
// the larger sums can swamp the new ones. As in nonNegSumState, both sums are
// therefore also recomputed whenever the volume sum, or the magnitude of the
// price-volume sum, falls below m2DropLimit of its peak since the last exact
// computation. Recomputing both together keeps their ratio accurate, and a
// window that traded nothing sums to exactly 0.
type VWAPPeriod struct {
	pv      *smaState // typical price × volume, pushed in step with vol
	vol     *smaState // volume
	pvPeak  float64   // the largest |Σ pv| since its last exact computation
	volPeak float64   // the largest Σ vol since its last exact computation
}

// NewVWAPPeriod constructs a VWAPPeriod with the given period.
func NewVWAPPeriod(period int) (*VWAPPeriod, error) {
	if period <= 0 {
		return nil, ErrInvalidPeriod
	}
	return &VWAPPeriod{pv: newSMAState(period), vol: newSMAState(period)}, nil
}

// MustNewVWAPPeriod constructs a VWAPPeriod or panics if the period is
// invalid.
func MustNewVWAPPeriod(period int) *VWAPPeriod {
	v, err := NewVWAPPeriod(period)
	if err != nil {
		panic(err)
	}
	return v
}

// Process consumes an OHLCV value and returns
//
//	Σ tp[i]·V[i] / Σ V[i],  tp = (H + L + C) / 3
//
// over the last period candles. The first value is ready on candle period,
// where it equals the cumulative VWAP of those candles. When the window's
// total volume is 0, it returns the current candle's typical price. Volumes
// must be ≥ 0.
func (w *VWAPPeriod) Process(v OHLCV) (float64, bool) {
	tp := (v.High + v.Low + v.Close) / 3
	w.pv.add(tp * v.Volume)
	if !w.vol.add(v.Volume) {
		return 0, false
	}
	pv, vol := w.pv.sum, w.vol.sum
	switch {
	case w.vol.idx == 0: // the window just filled, or wrapped and was recomputed
		w.pvPeak, w.volPeak = math.Abs(pv), vol
	case vol < w.volPeak*m2DropLimit || math.Abs(pv) < w.pvPeak*m2DropLimit:
		w.pv.recompute()
		w.vol.recompute()
		pv, vol = w.pv.sum, w.vol.sum
		w.pvPeak, w.volPeak = math.Abs(pv), vol
	default:
		w.pvPeak = max(w.pvPeak, math.Abs(pv))
		w.volPeak = max(w.volPeak, vol)
	}
	if vol <= 0 {
		return tp, true
	}
	return pv / vol, true
}

// Reset returns the VWAPPeriod to its newly constructed state.
func (w *VWAPPeriod) Reset() {
	w.pv.Reset()
	w.vol.Reset()
	w.pvPeak, w.volPeak = 0, 0
}
