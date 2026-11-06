# ta

Push-based technical analysis indicators in Go. Import as `ta "github.com/fedorwk/ta-go"`.

## Indicators

- Window primitives: Max, Min, Sum, Lag, TrueRange
- Statistics: Median, MAD, PercentRank, ZScore, RobustZ, RobustZPrior
- Moving averages: SMA, WMA, EMA, and MA with a selectable type (including Wilder's RMA)
- Momentum: RSI, MACD, Stochastic Oscillator (fast and full), Stochastic RSI, Williams %R, CCI, ROC, LogReturn, Momentum
- Trend: ADX with +DI/-DI
- Volatility: ATR, Bollinger Bands, Donchian Channels, StdDev, RealizedVol
- Volume: OBV, VWAP, MFI
- Bar features: CLV, EfficiencyRatio, RangeExpansion, RelVolume
- Regression: LinRegSlope, Regression, Correl

## Processor Interface

```go
type Processor[I any, O any] interface {
    Process(v I) (O, bool)
}
```

## SMA example

```go
s := ta.MustNewSMA(3)
for _, v := range []float64{1, 2, 3, 4, 5} {
    if avg, ok := s.Process(v); ok {
        fmt.Println(avg)
    }
}
```

## Composing indicators

`Chain` feeds one processor's output into another, and `Source` selects a price from a candle, so indicators that take plain values can run on candles:

```go
smoothedRSI := ta.Chain(ta.MustNewRSI(14), ta.MustNewSMA(5))
typicalSMA := ta.Chain(ta.HLC3, ta.MustNewSMA(20)) // Processor[ta.OHLC, float64]
```

`Lag` delays a series, so chaining it in front of a window processor computes over the values before the current one. For example, `ta.Chain(ta.MustNewLag(1), ta.MustNewMax(20))` is the highest of the 20 values before the current one.

`Apply` runs a processor over a slice and returns the ready outputs.

Indicators keep internal state and are not safe for concurrent use.

## Moving average types

Indicators that smooth with a configurable moving average take an `MAType`: `MATypeSMA`, `MATypeEMA`, `MATypeWMA`, or `MATypeRMA` (Wilder's smoothing). Each one emits its first value after `period` inputs. `NewMA` builds a standalone moving average of any of these types, which makes the type a runtime choice and is the way to use Wilder's RMA on its own.

```go
rma := ta.MustNewMA(14, ta.MATypeRMA)
slow := ta.MustNewStochFull(14, 3, 3, ta.MATypeSMA) // %K smoothed over 3, %D over 3
fast := ta.MustNewStoch(14, 3)                      // same as NewStochFull(14, 1, 3, MATypeSMA)
emaBands := ta.MustNewBollingerMA(20, 2, ta.MATypeEMA)
```

Bollinger Bands with a non-SMA middle band still take their width from the standard deviation around the simple mean, as TA-Lib's BBANDS does.

## Channels

Bollinger Bands and Donchian Channels both return `Bands`, which gives `PercentB`, the price's position in the channel ((price−Lower)/(Upper−Lower)), and `Bandwidth` ((Upper−Lower)/Middle). Bollinger uses the input value as the price, and Donchian uses the close. `PercentB` is 0.5 when the bands coincide, and `Bandwidth` is 0 when `Middle` is 0.

## Resetting

Every indicator has `Reset()`, which returns it to its newly constructed state without reallocating. Call `Reset` on `VWAP` at each session boundary to get a session (for example, daily) VWAP:

```go
vwap := ta.MustNewVWAP()
for _, c := range candles {
    if newSession(c) {
        vwap.Reset()
    }
    v, _ := vwap.Process(c)
    // ...
}
```

The processor that `Chain` returns also has a `Reset` method, which resets each stage that has one. Reach it with `p.(interface{ Reset() })`.

## Z-scores

`ZScore` scores each value against the mean and population standard deviation of its window. `RobustZ` and `RobustZPrior` use the median and the MAD scaled by 1.4826 instead, so one outlier barely moves the baseline:

- `RobustZ(period)` includes the current value in its window and is ready from input `period`.
- `RobustZPrior(period)` scores the current value against the `period` values before it, so it never affects its own baseline. It is ready from input `period+1`. Use it to detect spikes.

Both robust variants clamp their output to ±`RobustZCap` (10). When the MAD is 0, which happens whenever more than half the window is equal, a value scores 0 if it equals the median and ±`RobustZCap` otherwise.

## Returns and volatility

`LogReturn(period)` is ln(x / x₋period), the continuously compounded counterpart of `ROC`, and is ready from input `period+1`. `RealizedVol(period)` is √(Σ r²) over the last `period` one-step log returns, ready from input `period+1`. It is neither annualised nor de-meaned. To annualise, scale by √(periods per year / `period`). Both treat a return that involves a value ≤ 0 as 0.

## Bar features

- `CLV` is the close location value, (C − L) / (H − L), which runs from 0 at the low to 1 at the high. It is ready on every candle. It returns 0.5 when H = L and clamps a close outside the range. TA-Lib's signed form is `2·CLV − 1`, and multiplying that by volume gives one step of the AD line.
- `EfficiencyRatio(period)` is |x − x₋period| divided by the sum of |Δx| over those `period` steps. It is 1 for a one-way move, near 0 for chop, and 0 for a flat window. It is ready from input `period+1`.
- `RangeExpansion(period)` is the candle's true range divided by the median true range of the `period` candles before it. The first true range comes on the 2nd candle, so it is ready from candle `period+2`.
- `RelVolume(period)` is the candle's volume divided by the median volume of the `period` candles before it. It is ready from candle `period+1`.

The baselines of `RangeExpansion` and `RelVolume` leave out the current candle, so a spike cannot dilute its own baseline. Both return 0 when the baseline is 0.

## Regression

`Regression(period)` and `Correl(period)` take a `Pair{X, Y}` per input, for example a price change and the order-flow imbalance of the same bar:

- `Regression` returns a `RegressionValue`: the `Slope` and `Intercept` of the ordinary least squares fit of Y on X over the last `period` pairs, and its `R2`. This is a plain OLS fit of the raw values, not TA-Lib's BETA, which regresses the percentage changes of two price series.
- `Correl` returns the Pearson correlation of X and Y over the window, as TA-Lib's CORREL does.
- `LinRegSlope(period)` takes plain values and returns the slope of the least squares line through the last `period` of them against their position, as TA-Lib's LINEARREG_SLOPE does.

All three are ready from input `period`. When every X in the window is equal, the slope is undefined, so `Regression` returns slope 0, intercept ȳ and R² 0. When every Y is equal, the slope and R² are 0. `Correl` returns 0 when either side is constant, and `LinRegSlope` returns 0 for a flat window. The running sums are anchored at the window means and recomputed exactly as the window wraps, so the results stay accurate for prices near 1e5 and over long streams.

## Reference tests

Indicator tests compare outputs against [`github.com/markcheno/go-talib`](https://github.com/markcheno/go-talib), a pure-Go TA-Lib implementation. There are three exceptions:

- MACD: go-talib's MACD seeds the signal line with zeros, so its test uses a port of TA-Lib C's algorithm instead.
- Full Stochastic with EMA smoothing and a slow-%K period above 1: go-talib's `Stoch` computes %D over the zero-padded slow-%K buffer, which corrupts the EMA seed. The test averages only the defined slow-%K values, as TA-Lib C does.
- `MATypeRMA`: TA-Lib has no Wilder moving average type, so tests of it compare against a direct implementation, combined with go-talib's fast %K and standard deviation.
