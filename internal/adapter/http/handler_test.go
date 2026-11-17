package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ndresdavila/sezzle-calculator-api/internal/application/service"
	"github.com/ndresdavila/sezzle-calculator-api/internal/domain"
)

func TestAPI(t *testing.T) {
	router := NewRouter(service.New(), "*")

	t.Run("health", func(t *testing.T) {
		response := perform(router, http.MethodGet, "/health", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status %d", response.Code)
		}
		var body map[string]string
		mustDecode(t, response, &body)
		if body["status"] != "ok" {
			t.Fatalf("body %#v", body)
		}
	})

	t.Run("lists operations", func(t *testing.T) {
		response := perform(router, http.MethodGet, "/api/v1/operations", "")
		if response.Code != http.StatusOK {
			t.Fatalf("status %d", response.Code)
		}
		var body operationsResponse
		mustDecode(t, response, &body)
		names := map[string]bool{}
		for _, operation := range body.Operations {
			names[operation.Name] = true
		}
		for _, name := range []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "percent", "negate"} {
			if !names[name] {
				t.Fatalf("missing %s in %#v", name, body.Operations)
			}
		}
	})

	t.Run("adds decimal strings exactly", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/calculate", `{"operation":"add","operands":["0.1","0.2"]}`)
		if response.Code != http.StatusOK {
			t.Fatalf("status %d body %s", response.Code, response.Body.String())
		}
		var body calculateResponse
		mustDecode(t, response, &body)
		if body.Result != "0.3" {
			t.Fatalf("got %s", body.Result)
		}
	})

	t.Run("accepts json numbers", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/calculate", `{"operation":"divide","operands":[7,2]}`)
		var body calculateResponse
		mustDecode(t, response, &body)
		if response.Code != http.StatusOK || body.Result != "3.5" {
			t.Fatalf("status %d result %s", response.Code, body.Result)
		}
	})

	t.Run("evaluates the reference expression", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/evaluate", `{"expression":"346+74*24"}`)
		var body evaluateResponse
		mustDecode(t, response, &body)
		if response.Code != http.StatusOK || body.Result != "2122" {
			t.Fatalf("status %d body %#v", response.Code, body)
		}
	})

	t.Run("overflow", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/evaluate", `{"expression":"10^70"}`)
		assertError(t, response, http.StatusUnprocessableEntity, domain.CodeOverflow)
	})

	t.Run("underflow", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/evaluate", `{"expression":"(-10)^-70"}`)
		assertError(t, response, http.StatusUnprocessableEntity, domain.CodeUnderflow)
	})

	t.Run("division by zero", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/evaluate", `{"expression":"1/0"}`)
		assertError(t, response, http.StatusUnprocessableEntity, domain.CodeDivisionByZero)
	})

	t.Run("negative square root", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/calculate", `{"operation":"sqrt","operands":["-4"]}`)
		assertError(t, response, http.StatusUnprocessableEntity, domain.CodeNegativeSqrt)
	})

	t.Run("unknown operation", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/calculate", `{"operation":"mod","operands":["1","2"]}`)
		assertError(t, response, http.StatusBadRequest, domain.CodeUnknownOperation)
	})

	t.Run("wrong arity", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/calculate", `{"operation":"add","operands":["1"]}`)
		assertError(t, response, http.StatusBadRequest, domain.CodeInvalidOperandCount)
	})

	t.Run("invalid operand", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/calculate", `{"operation":"add","operands":["nope",2]}`)
		assertError(t, response, http.StatusBadRequest, domain.CodeInvalidNumber)
	})

	t.Run("unknown field", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/evaluate", `{"expression":"1+1","extra":true}`)
		assertError(t, response, http.StatusBadRequest, domain.CodeInvalidJSON)
	})

	t.Run("empty expression", func(t *testing.T) {
		response := perform(router, http.MethodPost, "/api/v1/evaluate", `{"expression":"  "}`)
		assertError(t, response, http.StatusBadRequest, domain.CodeInvalidExpression)
	})

	t.Run("wrong method", func(t *testing.T) {
		response := perform(router, http.MethodGet, "/api/v1/evaluate", "")
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status %d", response.Code)
		}
	})

	t.Run("cors preflight", func(t *testing.T) {
		response := perform(router, http.MethodOptions, "/api/v1/evaluate", "")
		if response.Code != http.StatusNoContent {
			t.Fatalf("status %d", response.Code)
		}
		if response.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("cors header %q", response.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

func TestCalculateTimeout(t *testing.T) {
	handler := NewHandler(service.New())
	request := httptest.NewRequest(http.MethodPost, "/api/v1/calculate", strings.NewReader(`{"operation":"add","operands":["1","2"]}`))
	ctx, cancel := context.WithCancel(request.Context())
	cancel()
	request = request.WithContext(ctx)
	response := httptest.NewRecorder()
	handler.calculate(response, request)
	assertError(t, response, http.StatusRequestTimeout, "TIMEOUT")
}

func perform(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func mustDecode(t *testing.T, response *httptest.ResponseRecorder, dest any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), dest); err != nil {
		t.Fatalf("decode %s: %v", response.Body.String(), err)
	}
}

func assertError(t *testing.T, response *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("status %d body %s", response.Code, response.Body.String())
	}
	var body errorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != code || body.Error.Message == "" {
		t.Fatalf("error %#v", body.Error)
	}
}
