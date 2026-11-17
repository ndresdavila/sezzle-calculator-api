package domain

import (
	"strings"

	"github.com/shopspring/decimal"
)

const (
	// resultScale is the number of decimal places kept on a finished result.
	resultScale = 12
	// maxMagnitude rejects results whose order of magnitude cannot fit a calculator display.
	maxMagnitude    int32 = 60
	maxNumberLength       = 100
)

func init() {
	// High intermediate precision lets identities such as (1/3)*3 snap back to 1.
	decimal.DivisionPrecision = 34
}

// ParseDecimal accepts a plain or scientific decimal string.
func ParseDecimal(raw string) (decimal.Decimal, error) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > maxNumberLength {
		return decimal.Zero, ErrInvalidNumber
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return decimal.Zero, ErrInvalidNumber
	}
	return parsed, nil
}

// Normalize rounds a finished result and snaps values that are a rounding
// error away from an integer, so (1/3)*3 is "1" and 0.1+0.2 is "0.3".
func Normalize(value decimal.Decimal) (decimal.Decimal, error) {
	if err := magnitudeError(value); err != nil {
		return decimal.Zero, err
	}
	rounded := value.Round(resultScale)
	nearest := rounded.Round(0)
	if rounded.Sub(nearest).Abs().LessThan(decimal.New(1, -9)) {
		rounded = nearest
	}
	if err := magnitudeError(rounded); err != nil {
		return decimal.Zero, err
	}
	return rounded, nil
}

// magnitudeError rejects values whose size cannot fit the calculator.
// Sign does not matter: a huge negative is overflow, a tiny negative is underflow.
func magnitudeError(value decimal.Decimal) error {
	if value.IsZero() {
		return nil
	}
	order := int32(value.NumDigits()) + value.Exponent()
	if order > maxMagnitude {
		return ErrOverflow
	}
	if order < -maxMagnitude {
		return ErrUnderflow
	}
	return nil
}
