package domain

// Error is a client-facing failure with a stable code. Adapters map the
// code to an HTTP status without inspecting message text.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

const (
	CodeDivisionByZero      = "DIVISION_BY_ZERO"
	CodeNegativeSqrt        = "NEGATIVE_SQRT"
	CodeInvalidPower        = "INVALID_POWER"
	CodeOverflow            = "OVERFLOW"
	CodeInvalidExpression   = "INVALID_EXPRESSION"
	CodeInvalidNumber       = "INVALID_NUMBER"
	CodeInvalidOperandCount = "INVALID_OPERAND_COUNT"
	CodeUnknownOperation    = "UNKNOWN_OPERATION"
	CodeInvalidJSON         = "INVALID_JSON"
)

var (
	ErrDivisionByZero      = &Error{Code: CodeDivisionByZero, Message: "Cannot divide by zero"}
	ErrNegativeSqrt        = &Error{Code: CodeNegativeSqrt, Message: "Cannot take the square root of a negative number"}
	ErrInvalidPower        = &Error{Code: CodeInvalidPower, Message: "That power is not defined for these numbers"}
	ErrZeroPower           = &Error{Code: CodeInvalidPower, Message: "Zero cannot be raised to the power of zero"}
	ErrNegativePower       = &Error{Code: CodeInvalidPower, Message: "A negative number cannot be raised to a fractional power"}
	ErrOverflow            = &Error{Code: CodeOverflow, Message: "The result is too large to represent"}
	ErrInvalidExpression   = &Error{Code: CodeInvalidExpression, Message: "The expression is invalid"}
	ErrInvalidNumber       = &Error{Code: CodeInvalidNumber, Message: "One of the numbers is invalid"}
	ErrInvalidOperandCount = &Error{Code: CodeInvalidOperandCount, Message: "Wrong number of operands for this operation"}
	ErrUnknownOperation    = &Error{Code: CodeUnknownOperation, Message: "Unknown operation"}
	ErrInvalidJSON         = &Error{Code: CodeInvalidJSON, Message: "Request body is invalid"}
)
