package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cashfac-test/internal/config"
	httptransport "cashfac-test/internal/transport/http"
	"cashfac-test/internal/usecase"
	"cashfac-test/internal/usecase/rewriter"
	"cashfac-test/internal/usecase/source"
	"cashfac-test/internal/usecase/storage"
)

type App struct {
	server *http.Server
}

func New() (*App, error) {
	cfg := config.Load()

	newsRepo := storage.NewInMemoryNewsRepository()
	rewriterClient := rewriter.NewNoopClient()
	sourceClient := source.NewStubClient()

	newsUseCase := usecase.NewNewsUseCase(newsRepo, sourceClient, rewriterClient)
	handler := httptransport.NewHandler(newsUseCase)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           handler.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{server: server}, nil
}

func (a *App) Run() error {
	log.Printf("HTTP server listening on %s", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
