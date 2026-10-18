package ta

import "errors"

var (
	// ErrInvalidPeriod is returned when an indicator period is <= 0.
	ErrInvalidPeriod = errors.New("period must be greater than 0")
	// ErrInvalidMACDPeriods is returned when MACD periods are invalid.
	ErrInvalidMACDPeriods = errors.New("macd periods must satisfy 0 < fast < slow and signal > 0")
	// ErrInvalidStdDev is returned when a standard deviation multiplier is <= 0.
	ErrInvalidStdDev = errors.New("stddev must be greater than 0")
	// ErrInvalidMAType is returned when an MAType is not one of the defined constants.
	ErrInvalidMAType = errors.New("invalid moving average type")
)
