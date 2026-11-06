package ta

// equalRun tracks how many of the most recent values equal the last one.
type equalRun struct {
	last float64
	n    int
}

// push records v and returns the length of the trailing run of values equal
// to v, capped at limit.
func (r *equalRun) push(v float64, limit int) int {
	if r.n > 0 && v == r.last {
		r.n = min(r.n+1, limit)
	} else {
		r.n = 1
	}
	r.last = v
	return r.n
}

// reset forgets every value.
func (r *equalRun) reset() {
	r.n = 0
}

// pairWindowState maintains the means, variances and covariance of a sliding
// window of pairs.
//
// Running sums of x, x² and so on lose every significant digit of a variance
// when the values are large relative to their spread, as prices near 1e5
// are. The sums are therefore taken of dx = x − ax and dy = y − ay, where the
// anchors ax and ay are the window means as of the last exact computation,
// so the terms stay of the order of the spread. Each time the ring wraps,
// the anchors and sums are recomputed exactly from the window, so rounding
// error cannot accumulate over long streams.
//
// As in varianceState, each update rounds a sum by about one ulp of its
// current size, so when the spread of X or Y collapses, the residue of the
// larger sums can swamp the new centred sum. The sums are therefore also
// recomputed whenever the centred sum of squares of X or Y falls below
// m2DropLimit of its peak since the last exact computation. This needs the
// spread to collapse by that factor, so it is rare and the cost stays
// amortized O(1).
//
// A side of the window whose values are all equal has its mean set to that
// value exactly and its centred sums to exactly 0, rather than to rounding
// residue, which would otherwise make a ratio such as the slope noise of
// order 1.
type pairWindowState struct {
	period int
	n      float64 // period as a float64
	window []Pair
	count  int
	idx    int

	ax, ay                float64 // anchors
	sx, sy, sxx, syy, sxy float64 // sums of dx, dy, dx², dy², dx·dy
	peakX, peakY          float64 // the largest centred sums of squares since the last exact computation

	runX, runY equalRun
}

func newPairWindowState(period int) *pairWindowState {
	return &pairWindowState{
		period: period,
		n:      float64(period),
		window: make([]Pair, period),
	}
}

// push adds a pair and reports whether the window is full.
func (s *pairWindowState) push(p Pair) bool {
	s.runX.push(p.X, s.period)
	s.runY.push(p.Y, s.period)

	if s.count < s.period {
		s.window[s.count] = p
		s.count++
		if s.count < s.period {
			return false
		}
		s.recompute()
		return true
	}

	old := s.window[s.idx]
	s.window[s.idx] = p
	s.idx++
	if s.idx == s.period {
		s.idx = 0
		s.recompute()
		return true
	}

	ox, oy := old.X-s.ax, old.Y-s.ay
	nx, ny := p.X-s.ax, p.Y-s.ay
	s.sx += nx - ox
	s.sy += ny - oy
	s.sxx += nx*nx - ox*ox
	s.syy += ny*ny - oy*oy
	s.sxy += nx*ny - ox*oy

	cxx, cyy := s.rawCxx(), s.rawCyy()
	if cxx < s.peakX*m2DropLimit || cyy < s.peakY*m2DropLimit {
		s.recompute()
		return true
	}
	s.peakX = max(s.peakX, cxx)
	s.peakY = max(s.peakY, cyy)
	return true
}

// recompute sets the anchors to the window means, or to the common value of
// a side whose values are all equal, and the sums exactly from the window.
func (s *pairWindowState) recompute() {
	if s.flatX() {
		s.ax = s.runX.last
	} else {
		sum := 0.0
		for _, p := range s.window {
			sum += p.X
		}
		s.ax = sum / s.n
	}
	if s.flatY() {
		s.ay = s.runY.last
	} else {
		sum := 0.0
		for _, p := range s.window {
			sum += p.Y
		}
		s.ay = sum / s.n
	}

	s.sx, s.sy, s.sxx, s.syy, s.sxy = 0, 0, 0, 0, 0
	for _, p := range s.window {
		dx, dy := p.X-s.ax, p.Y-s.ay
		s.sx += dx
		s.sy += dy
		s.sxx += dx * dx
		s.syy += dy * dy
		s.sxy += dx * dy
	}
	s.peakX, s.peakY = s.rawCxx(), s.rawCyy()
}

// rawCxx and rawCyy return the centred sums of squares as the running sums
// give them, before any clamping.
func (s *pairWindowState) rawCxx() float64 { return s.sxx - s.sx*s.sx/s.n }
func (s *pairWindowState) rawCyy() float64 { return s.syy - s.sy*s.sy/s.n }

// flatX and flatY report whether every X, or every Y, in the window is equal.
func (s *pairWindowState) flatX() bool { return s.runX.n == s.period }
func (s *pairWindowState) flatY() bool { return s.runY.n == s.period }

// meanX returns the mean of X over the window.
func (s *pairWindowState) meanX() float64 {
	if s.flatX() {
		return s.runX.last
	}
	return s.ax + s.sx/s.n
}

// meanY returns the mean of Y over the window.
func (s *pairWindowState) meanY() float64 {
	if s.flatY() {
		return s.runY.last
	}
	return s.ay + s.sy/s.n
}

// cxx returns Σ(x − x̄)² over the window, which is ≥ 0.
func (s *pairWindowState) cxx() float64 {
	if s.flatX() {
		return 0
	}
	return max(s.rawCxx(), 0)
}

// cyy returns Σ(y − ȳ)² over the window, which is ≥ 0.
func (s *pairWindowState) cyy() float64 {
	if s.flatY() {
		return 0
	}
	return max(s.rawCyy(), 0)
}

// cxy returns Σ(x − x̄)(y − ȳ) over the window.
func (s *pairWindowState) cxy() float64 {
	if s.flatX() || s.flatY() {
		return 0
	}
	return s.sxy - s.sx*s.sy/s.n
}

// Reset empties the window.
func (s *pairWindowState) Reset() {
	s.count, s.idx = 0, 0
	s.ax, s.ay = 0, 0
	s.sx, s.sy, s.sxx, s.syy, s.sxy = 0, 0, 0, 0, 0
	s.peakX, s.peakY = 0, 0
	s.runX.reset()
	s.runY.reset()
}
