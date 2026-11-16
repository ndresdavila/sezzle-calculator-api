package port

import (
	"context"

	"github.com/ndresdavila/sezzle-calculator-api/internal/domain"
	"github.com/shopspring/decimal"
)

// Calculator is the inbound port used by delivery adapters.
// A new transport (gRPC, CLI, queue) depends on this interface,
// not on the HTTP handler.
type Calculator interface {
	Calculate(ctx context.Context, operation domain.Operation, operands []decimal.Decimal) (decimal.Decimal, error)
	Evaluate(ctx context.Context, expression string) (decimal.Decimal, error)
	Operations(ctx context.Context) []domain.OperationInfo
}
