package utils

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/shopspring/decimal"
)

// decimalCache provides a cache for common decimal values to reduce allocations
var decimalCache = struct {
	sync.RWMutex
	values map[float64]decimal.Decimal
}{values: make(map[float64]decimal.Decimal, 100)}

// FloatToWire converts a float to a precise string representation
func FloatToWire(x float64) (string, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return "", fmt.Errorf("invalid float value: %v", x)
	}

	d, err := FloatToDecimal(x, DefaultDecimalPlaces)
	if err != nil {
		return "", err
	}

	return d.String(), nil
}

// FloatToDecimal converts a float to a decimal.Decimal with the specified precision
func FloatToDecimal(x float64, places int) (decimal.Decimal, error) {
	decimalCache.RLock()
	d, ok := decimalCache.values[x]
	decimalCache.RUnlock()

	if ok {
		return d, nil
	}

	rounded := fmt.Sprintf("%.*f", places, x)
	roundedFloat, err := strconv.ParseFloat(rounded, 64)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("parsing rounded float: %w", err)
	}

	if math.Abs(roundedFloat-x) >= PrecisionThreshold {
		return decimal.Decimal{}, fmt.Errorf("%w: %.12f vs %.12f", ErrPrecisionLoss, x, roundedFloat)
	}

	if rounded[0] == '-' && roundedFloat == 0 {
		rounded = rounded[1:]
	}

	d, err = decimal.NewFromString(rounded)
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("creating decimal: %w", err)
	}

	if len(decimalCache.values) < 1000 && math.Abs(x) < 1000000 {
		decimalCache.Lock()
		decimalCache.values[x] = d
		decimalCache.Unlock()
	}

	return d, nil
}

// FloatToIntForHashing converts a float to an integer with 8 decimal places
func FloatToIntForHashing(x float64) (int64, error) {
	return FloatToInt(x, DefaultDecimalPlaces)
}

// FloatToUSDInt converts a float to an integer with 6 decimal places for USD values
func FloatToUSDInt(x float64) (int64, error) {
	return FloatToInt(x, USDDecimalPlaces)
}

// FloatToInt converts a float to an integer with specified decimal places
func FloatToInt(x float64, places int) (int64, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, fmt.Errorf("invalid float value: %v", x)
	}

	scale := math.Pow10(places)
	withDecimals := x * scale

	if math.Abs(math.Round(withDecimals)-withDecimals) >= 1e-3 {
		return 0, fmt.Errorf("%w: %v would lose precision at %d decimal places",
			ErrPrecisionLoss, x, places)
	}

	result := int64(math.Round(withDecimals))
	if float64(result) != math.Round(withDecimals) {
		return 0, fmt.Errorf("integer overflow converting %v to int64", withDecimals)
	}

	return result, nil
}

// SafeFloat64 attempts to convert a string to a float64 with error handling
func SafeFloat64(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string cannot be converted to float64")
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing float: %w", err)
	}

	return f, nil
}

// DecimalToFloat64 safely converts a decimal.Decimal to float64, reporting errors
// if precision loss would occur beyond the specified tolerance
func DecimalToFloat64(d decimal.Decimal) (float64, error) {
	f := d.InexactFloat64()

	backToDecimal := decimal.NewFromFloat(f)
	if !backToDecimal.Equal(d) {
		diff := backToDecimal.Sub(d).Abs()
		if diff.GreaterThan(decimal.NewFromFloat(PrecisionThreshold)) {
			return f, fmt.Errorf("%w: decimal %s cannot be exactly represented as float64",
				ErrPrecisionLoss, d.String())
		}
	}

	return f, nil
}

// RoundFloat64 rounds a float64 to the specified number of decimal places
func RoundFloat64(x float64, places int) float64 {
	scale := math.Pow10(places)
	return math.Round(x*scale) / scale
}
