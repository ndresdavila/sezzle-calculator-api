package domain

import "github.com/shopspring/decimal"

// Operation is the public name of a single calculator function.
type Operation string

const (
	OpAdd      Operation = "add"
	OpSubtract Operation = "subtract"
	OpMultiply Operation = "multiply"
	OpDivide   Operation = "divide"
	OpPower    Operation = "power"
	OpSqrt     Operation = "sqrt"
	OpPercent  Operation = "percent"
	OpNegate   Operation = "negate"
)

// OperationInfo is the discoverable description of an operation.
// Adding a feature means adding one registry entry; the HTTP API lists it
// without further changes.
type OperationInfo struct {
	Name        Operation
	Arity       int
	Description string
}

type operation struct {
	info  OperationInfo
	apply func(operands []decimal.Decimal) (decimal.Decimal, error)
}

// registry is the single extension point for unary and binary operations.
var registry = []operation{
	{OperationInfo{OpAdd, 2, "Adds two numbers"}, add},
	{OperationInfo{OpSubtract, 2, "Subtracts the second number from the first"}, subtract},
	{OperationInfo{OpMultiply, 2, "Multiplies two numbers"}, multiply},
	{OperationInfo{OpDivide, 2, "Divides the first number by the second"}, divide},
	{OperationInfo{OpPower, 2, "Raises the first number to the power of the second"}, power},
	{OperationInfo{OpSqrt, 1, "Returns the square root of a non-negative number"}, sqrt},
	{OperationInfo{OpPercent, 1, "Divides a number by 100"}, percent},
	{OperationInfo{OpNegate, 1, "Changes the sign of a number"}, negate},
}

// Operations returns the catalog in a stable order.
func Operations() []OperationInfo {
	infos := make([]OperationInfo, len(registry))
	for i, entry := range registry {
		infos[i] = entry.info
	}
	return infos
}

// Apply runs one registered operation and normalizes the result.
func Apply(name Operation, operands []decimal.Decimal) (decimal.Decimal, error) {
	entry, ok := find(name)
	if !ok {
		return decimal.Zero, ErrUnknownOperation
	}
	if len(operands) != entry.info.Arity {
		return decimal.Zero, &Error{
			Code:    CodeInvalidOperandCount,
			Message: string(name) + " received the wrong number of operands",
		}
	}
	result, err := entry.apply(operands)
	if err != nil {
		return decimal.Zero, err
	}
	return Normalize(result)
}

func find(name Operation) (operation, bool) {
	for _, entry := range registry {
		if entry.info.Name == name {
			return entry, true
		}
	}
	return operation{}, false
}

func add(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Add(operands[1]), nil
}

func subtract(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Sub(operands[1]), nil
}

func multiply(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Mul(operands[1]), nil
}

func divide(operands []decimal.Decimal) (decimal.Decimal, error) {
	if operands[1].IsZero() {
		return decimal.Zero, ErrDivisionByZero
	}
	return operands[0].Div(operands[1]), nil
}

func percent(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Div(decimal.NewFromInt(100)), nil
}

func negate(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Neg(), nil
}

func sqrt(operands []decimal.Decimal) (decimal.Decimal, error) {
	return squareRoot(operands[0])
}

func power(operands []decimal.Decimal) (decimal.Decimal, error) {
	return raise(operands[0], operands[1])
}

func squareRoot(value decimal.Decimal) (decimal.Decimal, error) {
	if value.IsNegative() {
		return decimal.Zero, ErrNegativeSqrt
	}
	if value.IsZero() {
		return decimal.Zero, nil
	}

	two := decimal.NewFromInt(2)
	guess := value.Div(two)
	if guess.IsZero() {
		guess = value
	}
	epsilon := decimal.New(1, -18)
	for i := 0; i < 80; i++ {
		next := guess.Add(value.Div(guess)).Div(two)
		if next.Sub(guess).Abs().LessThan(epsilon) {
			return next, nil
		}
		guess = next
	}
	return guess, nil
}

func raise(base, exponent decimal.Decimal) (decimal.Decimal, error) {
	if exponent.Abs().GreaterThan(decimal.NewFromInt(256)) {
		if exponent.IsNegative() {
			return decimal.Zero, ErrUnderflow
		}
		return decimal.Zero, ErrOverflow
	}
	if base.IsZero() && exponent.IsZero() {
		return decimal.Zero, ErrZeroPower
	}
	if base.IsZero() && exponent.IsNegative() {
		return decimal.Zero, ErrDivisionByZero
	}
	if base.IsNegative() && !exponent.Equal(exponent.Truncate(0)) {
		return decimal.Zero, ErrNegativePower
	}
	result, err := base.PowWithPrecision(exponent, 16)
	if err != nil {
		return decimal.Zero, ErrOverflow
	}
	return result, nil
}
