package ta

// OHLC represents an open-high-low-close candle.
type OHLC struct {
	Open  float64
	High  float64
	Low   float64
	Close float64
}

// OHLCV represents an open-high-low-close-volume candle.
type OHLCV struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// Pair is one observation of two series, such as a price change and the
// order-flow imbalance over the same bar.
type Pair struct {
	X float64
	Y float64
}

// RegressionValue is the ordinary least squares fit of Y on X over a window:
// Y ≈ Slope·X + Intercept, with R2 the coefficient of determination.
type RegressionValue struct {
	Slope     float64
	Intercept float64
	R2        float64
}
