package httpadapter

import "encoding/json"

type calculateRequest struct {
	Operation string            `json:"operation"`
	Operands  []json.RawMessage `json:"operands"`
}

type evaluateRequest struct {
	Expression string `json:"expression"`
}

type calculateResponse struct {
	Operation string   `json:"operation"`
	Operands  []string `json:"operands"`
	Result    string   `json:"result"`
}

type evaluateResponse struct {
	Expression string `json:"expression"`
	Result     string `json:"result"`
}

type operationResponse struct {
	Name        string `json:"name"`
	Arity       int    `json:"arity"`
	Description string `json:"description"`
}

type operationsResponse struct {
	Operations []operationResponse `json:"operations"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}
