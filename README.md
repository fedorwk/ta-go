# ta

Push-based technical analysis indicators in Go. Import as `ta "github.com/fedorwk/ta-go"`.

## Indicators

- Moving averages: SMA, WMA, EMA, and MA with a selectable type (including Wilder's RMA)
- Momentum: RSI, MACD, Stochastic Oscillator (fast and full), Stochastic RSI, Williams %R, CCI, ROC, Momentum
- Trend: ADX with +DI/-DI
- Volatility: ATR, Bollinger Bands, Donchian Channels, StdDev
- Volume: OBV, VWAP, MFI

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

## Reference tests

Indicator tests compare outputs against [`github.com/markcheno/go-talib`](https://github.com/markcheno/go-talib), a pure-Go TA-Lib implementation. There are three exceptions:

- MACD: go-talib's MACD seeds the signal line with zeros, so its test uses a port of TA-Lib C's algorithm instead.
- Full Stochastic with EMA smoothing and a slow-%K period above 1: go-talib's `Stoch` computes %D over the zero-padded slow-%K buffer, which corrupts the EMA seed. The test averages only the defined slow-%K values, as TA-Lib C does.
- `MATypeRMA`: TA-Lib has no Wilder moving average type, so tests of it compare against a direct implementation, combined with go-talib's fast %K and standard deviation.
