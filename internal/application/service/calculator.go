// Package service implements the calculator use cases.
// It is deliberately thin: the arithmetic rules live in the domain,
// and this layer is where cross-cutting application policy (cancellation,
// future auditing, future history) belongs.
package service

import (
	"context"

	"github.com/ndresdavila/sezzle-calculator-api/internal/domain"
	"github.com/shopspring/decimal"
)

// Calculator is the application service behind the inbound port.
type Calculator struct{}

func New() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Calculate(ctx context.Context, operation domain.Operation, operands []decimal.Decimal) (decimal.Decimal, error) {
	if err := ctx.Err(); err != nil {
		return decimal.Zero, err
	}
	return domain.Apply(operation, operands)
}

func (c *Calculator) Evaluate(ctx context.Context, expression string) (decimal.Decimal, error) {
	if err := ctx.Err(); err != nil {
		return decimal.Zero, err
	}
	return domain.Evaluate(expression)
}

func (c *Calculator) Operations(ctx context.Context) []domain.OperationInfo {
	if err := ctx.Err(); err != nil {
		return nil
	}
	return domain.Operations()
}
