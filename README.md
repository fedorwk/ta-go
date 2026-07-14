# ta

Push-based technical analysis indicators in Go. Import as `ta "github.com/fedorwk/ta-go"`.

## Indicators

- SMA
- WMA
- EMA
- RSI
- MACD
- Bollinger Bands
- Stochastic Oscillator
- ATR
- OBV
- VWAP

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
    if avg, ok := s.Update(v); ok {
        fmt.Println(avg)
    }
}
```

## Benchmark tests

Indicator tests compare outputs against [`github.com/markcheno/go-talib`](https://github.com/markcheno/go-talib), a pure-Go TA-Lib implementation.
