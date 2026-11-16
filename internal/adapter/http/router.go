package httpadapter

import (
	"net/http"
	"time"

	"github.com/ndresdavila/sezzle-calculator-api/internal/application/port"
)

// NewRouter wires the HTTP routes and the edge middleware.
func NewRouter(calc port.Calculator, corsOrigins string) http.Handler {
	handler := NewHandler(calc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.health)
	mux.HandleFunc("GET /api/v1/operations", handler.operations)
	mux.HandleFunc("POST /api/v1/calculate", handler.calculate)
	mux.HandleFunc("POST /api/v1/evaluate", handler.evaluate)

	var httpHandler http.Handler = mux
	httpHandler = withTimeout(5*time.Second, httpHandler)
	httpHandler = cors(corsOrigins, httpHandler)
	httpHandler = logging(httpHandler)
	httpHandler = recoverPanics(httpHandler)
	return httpHandler
}
