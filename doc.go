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
// constructed state, for example to start a new VWAP session; VWAPPeriod is
// the rolling alternative, over the last period candles. Indicators that
// smooth with a configurable moving average take an MAType.
//
// Chain composes processors, and Source adapts OHLC candles to indicators
// that consume plain values. The window primitives Max, Min, Sum, Lag and
// TrueRange are building blocks for such chains; Lag in front of a window
// processor computes over the values before the current one.
//
// Median, MAD and PercentRank are rolling order statistics, kept on a sorted
// copy of the window. RobustZ and RobustZPrior build a z-score from the
// median and MAD in the same way, clamped to ±RobustZCap. None of them
// support NaN inputs. ZScore is the counterpart that uses the mean and
// standard deviation.
//
// LogReturn and RealizedVol work on log returns, and treat a return that
// involves a value ≤ 0 as 0.
//
// CLV, EfficiencyRatio, RangeExpansion and RelVolume describe the shape of
// bars: where a candle closed in its range, how directly a series moved, and
// how a candle's true range and volume compare with the median of the
// candles before it.
//
// Regression and Correl fit and correlate two series given as Pair values,
// and LinRegSlope fits a single series against its position. Regression is
// plain least squares on the raw values, not TA-Lib's BETA.
//
// Indicators are not safe for concurrent use.
package ta
