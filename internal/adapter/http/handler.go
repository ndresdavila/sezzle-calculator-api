package httpadapter

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/ndresdavila/sezzle-calculator-api/internal/application/port"
	"github.com/ndresdavila/sezzle-calculator-api/internal/domain"
	"github.com/shopspring/decimal"
)

// Handler is the HTTP adapter. It translates JSON and status codes;
// it does not implement arithmetic.
type Handler struct {
	calc port.Calculator
}

func NewHandler(calc port.Calculator) *Handler {
	return &Handler{calc: calc}
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) operations(w http.ResponseWriter, r *http.Request) {
	infos := h.calc.Operations(r.Context())
	payload := make([]operationResponse, 0, len(infos))
	for _, info := range infos {
		payload = append(payload, operationResponse{
			Name:        string(info.Name),
			Arity:       info.Arity,
			Description: info.Description,
		})
	}
	writeJSON(w, http.StatusOK, operationsResponse{Operations: payload})
}

func (h *Handler) calculate(w http.ResponseWriter, r *http.Request) {
	var request calculateRequest
	if err := decode(w, r, &request); err != nil {
		writeDomainError(w, err)
		return
	}

	operands := make([]decimal.Decimal, 0, len(request.Operands))
	rendered := make([]string, 0, len(request.Operands))
	for _, raw := range request.Operands {
		operand, err := parseOperand(raw)
		if err != nil {
			writeDomainError(w, err)
			return
		}
		operands = append(operands, operand)
		rendered = append(rendered, operand.String())
	}

	result, err := h.calc.Calculate(r.Context(), domain.Operation(request.Operation), operands)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, calculateResponse{
		Operation: request.Operation,
		Operands:  rendered,
		Result:    result.String(),
	})
}

func (h *Handler) evaluate(w http.ResponseWriter, r *http.Request) {
	var request evaluateRequest
	if err := decode(w, r, &request); err != nil {
		writeDomainError(w, err)
		return
	}
	expression := strings.TrimSpace(request.Expression)
	result, err := h.calc.Evaluate(r.Context(), expression)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evaluateResponse{
		Expression: expression,
		Result:     result.String(),
	})
}

func parseOperand(raw json.RawMessage) (decimal.Decimal, error) {
	raw = bytesTrim(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return decimal.Zero, domain.ErrInvalidNumber
	}
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return decimal.Zero, domain.ErrInvalidNumber
		}
		return domain.ParseDecimal(text)
	}
	if raw[0] == '-' || raw[0] == '+' || (raw[0] >= '0' && raw[0] <= '9') {
		return domain.ParseDecimal(string(raw))
	}
	return decimal.Zero, domain.ErrInvalidNumber
}

func bytesTrim(raw json.RawMessage) json.RawMessage {
	return json.RawMessage(strings.TrimSpace(string(raw)))
}

func decode(w http.ResponseWriter, r *http.Request, dest any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return domain.ErrInvalidJSON
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.ErrInvalidJSON
	}
	return nil
}

func writeDomainError(w http.ResponseWriter, err error) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		writeJSON(w, statusFor(domainErr.Code), errorResponse{
			Error: errorBody{Code: domainErr.Code, Message: domainErr.Message},
		})
		return
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		writeJSON(w, http.StatusRequestTimeout, errorResponse{
			Error: errorBody{Code: "TIMEOUT", Message: "The request timed out"},
		})
		return
	}
	writeJSON(w, http.StatusInternalServerError, errorResponse{
		Error: errorBody{Code: "INTERNAL", Message: "Something went wrong"},
	})
}

func statusFor(code string) int {
	switch code {
	case domain.CodeDivisionByZero, domain.CodeNegativeSqrt, domain.CodeInvalidPower, domain.CodeOverflow:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadRequest
	}
}
