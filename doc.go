// Package ta provides push-based technical analysis indicators.
//
// Each indicator is created with its parameters and fed values one at a time
// via its Update method. Update returns the computed value paired with a
// boolean that is false while the indicator is warming up (e.g. until a
// moving average window is full) and true thereafter. Indicators consuming
// OHLC data accept OHLC structs; indicators that need volume accept OHLCV
// structs.
package ta