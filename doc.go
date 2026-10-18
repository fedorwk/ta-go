// Package ta provides push-based technical analysis indicators.
//
// Each indicator is created with its parameters and fed values one at a time
// via its Process method. Process returns the computed value paired with a
// boolean that is false while the indicator is warming up (e.g. until a
// moving average window is full) and true thereafter. Indicators consuming
// OHLC data accept OHLC structs; indicators that need volume accept OHLCV
// structs.
//
// Every indicator has a Reset method that returns it to its newly
// constructed state, for example to start a new VWAP session. Indicators
// that smooth with a configurable moving average take an MAType.
//
// Chain composes processors, and Source adapts OHLC candles to indicators
// that consume plain values. Indicators are not safe for concurrent use.
package ta
