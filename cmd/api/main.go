package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/ndresdavila/sezzle-calculator-api/internal/adapter/http"
	"github.com/ndresdavila/sezzle-calculator-api/internal/application/service"
	"github.com/ndresdavila/sezzle-calculator-api/internal/platform/config"
)

func main() {
	cfg := config.FromEnv()
	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           httpadapter.NewRouter(service.New(), cfg.CORSOrigins),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("calculator api listening on %s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
