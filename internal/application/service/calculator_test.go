package service

import (
	"context"
	"testing"

	"github.com/ndresdavila/sezzle-calculator-api/internal/domain"
	"github.com/shopspring/decimal"
)

func TestCalculateDelegatesToDomain(t *testing.T) {
	calc := New()
	got, err := calc.Calculate(context.Background(), domain.OpAdd, []decimal.Decimal{
		decimal.NewFromInt(346),
		decimal.NewFromInt(1776),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "2122" {
		t.Fatalf("got %s", got.String())
	}
}

func TestEvaluateDelegatesToDomain(t *testing.T) {
	got, err := New().Evaluate(context.Background(), "346+74*24")
	if err != nil {
		t.Fatal(err)
	}
	if got.String() != "2122" {
		t.Fatalf("got %s", got.String())
	}
}

func TestOperationsAreDiscoverable(t *testing.T) {
	ops := New().Operations(context.Background())
	if len(ops) < 4 {
		t.Fatalf("expected the basic operations, got %d", len(ops))
	}
}

func TestCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New().Calculate(ctx, domain.OpAdd, nil); err == nil {
		t.Fatal("expected cancellation")
	}
	if _, err := New().Evaluate(ctx, "1+1"); err == nil {
		t.Fatal("expected cancellation")
	}
	if New().Operations(ctx) != nil {
		t.Fatal("expected no catalog once the context is canceled")
	}
}
