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
