package domain

import (
	"strings"
	"testing"

	"github.com/shopspring/decimal"
)

func TestApplyOperations(t *testing.T) {
	cases := []struct {
		name     string
		op       Operation
		operands []string
		want     string
	}{
		{name: "add", op: OpAdd, operands: []string{"0.1", "0.2"}, want: "0.3"},
		{name: "subtract", op: OpSubtract, operands: []string{"5", "2"}, want: "3"},
		{name: "multiply", op: OpMultiply, operands: []string{"74", "24"}, want: "1776"},
		{name: "divide", op: OpDivide, operands: []string{"7", "2"}, want: "3.5"},
		{name: "power", op: OpPower, operands: []string{"2", "10"}, want: "1024"},
		{name: "negative power", op: OpPower, operands: []string{"2", "-2"}, want: "0.25"},
		{name: "sqrt", op: OpSqrt, operands: []string{"9"}, want: "3"},
		{name: "percent", op: OpPercent, operands: []string{"50"}, want: "0.5"},
		{name: "negate", op: OpNegate, operands: []string{"4"}, want: "-4"},
		{name: "negate zero", op: OpNegate, operands: []string{"0"}, want: "0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Apply(tc.op, mustDecimals(t, tc.operands))
			if err != nil {
				t.Fatalf("Apply returned error: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %s, want %s", got.String(), tc.want)
			}
		})
	}
}

func TestEveryCatalogOperationSucceedsWithFours(t *testing.T) {
	for _, info := range Operations() {
		operands := make([]decimal.Decimal, info.Arity)
		for i := range operands {
			operands[i] = decimal.NewFromInt(4)
		}
		if _, err := Apply(info.Name, operands); err != nil {
			t.Fatalf("%s: %v", info.Name, err)
		}
	}
}

func TestApplyRejectsBadInput(t *testing.T) {
	cases := []struct {
		name     string
		op       Operation
		operands []string
		code     string
	}{
		{name: "divide by zero", op: OpDivide, operands: []string{"1", "0"}, code: CodeDivisionByZero},
		{name: "negative sqrt", op: OpSqrt, operands: []string{"-1"}, code: CodeNegativeSqrt},
		{name: "zero to zero", op: OpPower, operands: []string{"0", "0"}, code: CodeInvalidPower},
		{name: "negative fractional power", op: OpPower, operands: []string{"-2", "0.5"}, code: CodeInvalidPower},
		{name: "huge exponent", op: OpPower, operands: []string{"2", "999"}, code: CodeOverflow},
		{name: "wrong arity", op: OpAdd, operands: []string{"1"}, code: CodeInvalidOperandCount},
		{name: "unknown", op: "mod", operands: []string{"1", "2"}, code: CodeUnknownOperation},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Apply(tc.op, mustDecimals(t, tc.operands))
			assertCode(t, err, tc.code)
		})
	}
}

func TestParseDecimal(t *testing.T) {
	parsed, err := ParseDecimal(" 12.50 ")
	if err != nil || parsed.String() != "12.5" {
		t.Fatalf("got %s, %v", parsed, err)
	}
	if _, err := ParseDecimal(""); err == nil {
		t.Fatal("expected empty input to fail")
	}
	if _, err := ParseDecimal("nope"); err == nil {
		t.Fatal("expected text to fail")
	}
	if _, err := ParseDecimal(strings.Repeat("1", 101)); err == nil {
		t.Fatal("expected a huge literal to fail")
	}
}

func TestNormalizeSnapsNearIntegers(t *testing.T) {
	third, err := decimal.NewFromString("1")
	if err != nil {
		t.Fatal(err)
	}
	value, err := Normalize(third.Div(decimal.NewFromInt(3)).Mul(decimal.NewFromInt(3)))
	if err != nil {
		t.Fatal(err)
	}
	if value.String() != "1" {
		t.Fatalf("got %s", value.String())
	}
}

func mustDecimals(t *testing.T, raw []string) []decimal.Decimal {
	t.Helper()
	values := make([]decimal.Decimal, len(raw))
	for i, item := range raw {
		parsed, err := ParseDecimal(item)
		if err != nil {
			t.Fatalf("operand %s: %v", item, err)
		}
		values[i] = parsed
	}
	return values
}

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	domainErr, ok := err.(*Error)
	if !ok || domainErr.Code != code {
		t.Fatalf("got %v, want code %s", err, code)
	}
}
