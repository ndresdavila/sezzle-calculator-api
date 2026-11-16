package domain

import (
	"strings"
	"unicode/utf8"

	"github.com/shopspring/decimal"
)

const (
	maxExpressionLength = 500
	maxDepth            = 64
)

// value keeps a percent flag so "200+10%" can mean 200 plus 10 percent of 200,
// while a bare "10%" still means 0.1. Parentheses resolve the percent immediately.
type value struct {
	n               decimal.Decimal
	relativePercent bool
}

type parser struct {
	s     string
	i     int
	depth int
}

// Evaluate parses an arithmetic expression and returns a normalized result.
//
// Grammar, highest precedence first:
//
//	primary  = number | "(" expr ")" | "sqrt" "(" expr ")"
//	postfix  = primary { "%" }
//	unary    = ("+" | "-") unary | postfix
//	power    = unary { "^" power }          // right associative
//	term     = power { ("*" | "/") power }
//	expr     = term { ("+" | "-") term }
//
// A percent that is the direct right operand of + or - is relative to the
// left operand (200+10% = 220). Anywhere else, % divides that value by 100.
func Evaluate(expression string) (decimal.Decimal, error) {
	trimmed := strings.TrimSpace(expression)
	if trimmed == "" || len(trimmed) > maxExpressionLength {
		return decimal.Zero, ErrInvalidExpression
	}
	p := &parser{s: trimmed}
	parsed, err := p.parseExpr()
	if err != nil {
		return decimal.Zero, err
	}
	if p.peek() != 0 {
		return decimal.Zero, ErrInvalidExpression
	}
	return Normalize(materialize(parsed))
}

func materialize(v value) decimal.Decimal {
	if !v.relativePercent {
		return v.n
	}
	return v.n.Div(decimal.NewFromInt(100))
}

func (p *parser) enter() error {
	p.depth++
	if p.depth > maxDepth {
		return ErrInvalidExpression
	}
	return nil
}

func (p *parser) leave() { p.depth-- }

func (p *parser) parseExpr() (value, error) {
	if err := p.enter(); err != nil {
		return value{}, err
	}
	defer p.leave()

	left, err := p.parseTerm()
	if err != nil {
		return value{}, err
	}
	for {
		op := p.peek()
		if op != '+' && op != '-' && op != '−' {
			break
		}
		p.consume(op)
		right, err := p.parseTerm()
		if err != nil {
			return value{}, err
		}
		left, err = applyAddSub(op, left, right)
		if err != nil {
			return value{}, err
		}
	}
	return left, nil
}

func (p *parser) parseTerm() (value, error) {
	left, err := p.parsePower()
	if err != nil {
		return value{}, err
	}
	for {
		op := p.peek()
		if op != '*' && op != '/' && op != '×' && op != '÷' {
			break
		}
		p.consume(op)
		right, err := p.parsePower()
		if err != nil {
			return value{}, err
		}
		left, err = applyMulDiv(op, left, right)
		if err != nil {
			return value{}, err
		}
	}
	return left, nil
}

func (p *parser) parsePower() (value, error) {
	left, err := p.parseUnary()
	if err != nil {
		return value{}, err
	}
	if p.peek() != '^' {
		return left, nil
	}
	p.consume('^')
	right, err := p.parsePower()
	if err != nil {
		return value{}, err
	}
	result, err := raise(materialize(left), materialize(right))
	if err != nil {
		return value{}, err
	}
	return value{n: result}, nil
}

func (p *parser) parseUnary() (value, error) {
	op := p.peek()
	if op != '+' && op != '-' && op != '−' {
		return p.parsePostfix()
	}
	p.consume(op)
	inner, err := p.parseUnary()
	if err != nil {
		return value{}, err
	}
	if op == '+' {
		return inner, nil
	}
	inner.n = inner.n.Neg()
	return inner, nil
}

func (p *parser) parsePostfix() (value, error) {
	parsed, err := p.parsePrimary()
	if err != nil {
		return value{}, err
	}
	for p.peek() == '%' {
		p.consume('%')
		if parsed.relativePercent {
			parsed.n = parsed.n.Div(decimal.NewFromInt(100))
		}
		parsed.relativePercent = true
	}
	return parsed, nil
}

func (p *parser) parsePrimary() (value, error) {
	if err := p.enter(); err != nil {
		return value{}, err
	}
	defer p.leave()

	if p.peek() == '(' {
		p.consume('(')
		inner, err := p.parseExpr()
		if err != nil {
			return value{}, err
		}
		if !p.consume(')') {
			return value{}, ErrInvalidExpression
		}
		return value{n: materialize(inner)}, nil
	}
	if p.hasIdent("sqrt") {
		p.i += len("sqrt")
		if !p.consume('(') {
			return value{}, ErrInvalidExpression
		}
		inner, err := p.parseExpr()
		if err != nil {
			return value{}, err
		}
		if !p.consume(')') {
			return value{}, ErrInvalidExpression
		}
		root, err := squareRoot(materialize(inner))
		if err != nil {
			return value{}, err
		}
		return value{n: root}, nil
	}
	number, err := p.parseNumber()
	if err != nil {
		return value{}, err
	}
	return value{n: number}, nil
}

func (p *parser) parseNumber() (decimal.Decimal, error) {
	p.skip()
	start := p.i
	sawDigit := false
	for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
		sawDigit = true
		p.i++
	}
	if p.i < len(p.s) && p.s[p.i] == '.' {
		p.i++
		fraction := p.i
		for p.i < len(p.s) && p.s[p.i] >= '0' && p.s[p.i] <= '9' {
			sawDigit = true
			p.i++
		}
		if p.i == fraction {
			return decimal.Zero, ErrInvalidExpression
		}
	}
	if !sawDigit {
		return decimal.Zero, ErrInvalidExpression
	}
	parsed, err := decimal.NewFromString(p.s[start:p.i])
	if err != nil {
		return decimal.Zero, ErrInvalidNumber
	}
	return parsed, nil
}

func applyAddSub(op rune, left, right value) (value, error) {
	base := materialize(left)
	if right.relativePercent {
		delta, err := multiply([]decimal.Decimal{base, right.n})
		if err != nil {
			return value{}, err
		}
		delta, err = divide([]decimal.Decimal{delta, decimal.NewFromInt(100)})
		if err != nil {
			return value{}, err
		}
		var result decimal.Decimal
		if op == '+' {
			result, err = add([]decimal.Decimal{base, delta})
		} else {
			result, err = subtract([]decimal.Decimal{base, delta})
		}
		if err != nil {
			return value{}, err
		}
		return value{n: result}, nil
	}
	other := materialize(right)
	var (
		result decimal.Decimal
		err    error
	)
	if op == '+' {
		result, err = add([]decimal.Decimal{base, other})
	} else {
		result, err = subtract([]decimal.Decimal{base, other})
	}
	if err != nil {
		return value{}, err
	}
	return value{n: result}, nil
}

func applyMulDiv(op rune, left, right value) (value, error) {
	operands := []decimal.Decimal{materialize(left), materialize(right)}
	var (
		result decimal.Decimal
		err    error
	)
	if op == '/' || op == '÷' {
		result, err = divide(operands)
	} else {
		result, err = multiply(operands)
	}
	if err != nil {
		return value{}, err
	}
	return value{n: result}, nil
}

func (p *parser) hasIdent(word string) bool {
	p.skip()
	if !strings.HasPrefix(p.s[p.i:], word) {
		return false
	}
	next := p.i + len(word)
	if next < len(p.s) {
		r := p.s[next]
		if isIdentChar(r) {
			return false
		}
	}
	return true
}

func isIdentChar(r byte) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

func (p *parser) skip() {
	for p.i < len(p.s) {
		r, size := utf8.DecodeRuneInString(p.s[p.i:])
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return
		}
		p.i += size
	}
}

func (p *parser) peek() rune {
	p.skip()
	if p.i >= len(p.s) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(p.s[p.i:])
	return r
}

func (p *parser) consume(want rune) bool {
	if p.peek() != want {
		return false
	}
	_, size := utf8.DecodeRuneInString(p.s[p.i:])
	p.i += size
	return true
}
