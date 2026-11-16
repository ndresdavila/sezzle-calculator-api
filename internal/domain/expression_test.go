package domain

import (
	"strings"
	"testing"
)

func TestEvaluate(t *testing.T) {
	cases := []struct {
		expression string
		want       string
	}{
		{expression: "346+74*24", want: "2122"},
		{expression: " 346 + 74 * 24 ", want: "2122"},
		{expression: "346+74×24", want: "2122"},
		{expression: "8÷2", want: "4"},
		{expression: "10−3", want: "7"},
		{expression: "0.1+0.2", want: "0.3"},
		{expression: "2-5", want: "-3"},
		{expression: "8*3", want: "24"},
		{expression: "7/2", want: "3.5"},
		{expression: "10/4", want: "2.5"},
		{expression: "(1/3)*3", want: "1"},
		{expression: "50%", want: "0.5"},
		{expression: "200+10%", want: "220"},
		{expression: "200-10%", want: "180"},
		{expression: "200*10%", want: "20"},
		{expression: "200/10%", want: "2000"},
		{expression: "100+50%+10", want: "160"},
		{expression: "(10%)", want: "0.1"},
		{expression: "200+(10%)", want: "200.1"},
		{expression: "9%%", want: "0.0009"},
		{expression: "sqrt(9)", want: "3"},
		{expression: "sqrt(4)+6", want: "8"},
		{expression: "sqrt(9%)", want: "0.3"},
		{expression: "2^10", want: "1024"},
		{expression: "2^-2", want: "0.25"},
		{expression: "4^0.5", want: "2"},
		{expression: "2^3^2", want: "512"},
		{expression: "(2^10)%", want: "10.24"},
		{expression: "5+-3", want: "2"},
		{expression: "-5+2", want: "-3"},
		{expression: "5--3", want: "8"},
		{expression: "+5", want: "5"},
		{expression: "(2+3)*4", want: "20"},
		{expression: "200+-10%", want: "180"},
	}

	for _, tc := range cases {
		t.Run(tc.expression, func(t *testing.T) {
			got, err := Evaluate(tc.expression)
			if err != nil {
				t.Fatalf("Evaluate returned error: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %s, want %s", got.String(), tc.want)
			}
		})
	}
}

func TestEvaluateSqrtTwo(t *testing.T) {
	got, err := Evaluate("sqrt(2)")
	if err != nil {
		t.Fatal(err)
	}
	text := got.String()
	if !strings.HasPrefix(text, "1.41421356237") {
		t.Fatalf("got %s", text)
	}
}

func TestEvaluateErrors(t *testing.T) {
	cases := []struct {
		expression string
		code       string
	}{
		{expression: "", code: CodeInvalidExpression},
		{expression: "   ", code: CodeInvalidExpression},
		{expression: "5+", code: CodeInvalidExpression},
		{expression: "5**2", code: CodeInvalidExpression},
		{expression: "()", code: CodeInvalidExpression},
		{expression: "sqrt", code: CodeInvalidExpression},
		{expression: "5.", code: CodeInvalidExpression},
		{expression: strings.Repeat("1+", 300) + "1", code: CodeInvalidExpression},
		{expression: strings.Repeat("(", 40) + "1" + strings.Repeat(")", 40), code: CodeInvalidExpression},
		{expression: "1/0", code: CodeDivisionByZero},
		{expression: "10/0%", code: CodeDivisionByZero},
		{expression: "sqrt(-1)", code: CodeNegativeSqrt},
		{expression: "0^0", code: CodeInvalidPower},
		{expression: "(-2)^0.5", code: CodeInvalidPower},
	}

	for _, tc := range cases {
		t.Run(tc.expression, func(t *testing.T) {
			_, err := Evaluate(tc.expression)
			assertCode(t, err, tc.code)
		})
	}
}

func FuzzEvaluate(f *testing.F) {
	f.Add("346+74*24")
	f.Add("1/0")
	f.Add("sqrt(-1)")
	f.Fuzz(func(t *testing.T, expression string) {
		_, _ = Evaluate(expression)
	})
}
