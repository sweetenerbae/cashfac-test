package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cashfac-test/internal/config"
	"cashfac-test/internal/domain"
	httptransport "cashfac-test/internal/transport/http"
	"cashfac-test/internal/usecase"
	"cashfac-test/internal/usecase/rewriter"
	"cashfac-test/internal/usecase/source"
	"cashfac-test/internal/usecase/storage"
)

type App struct {
	server *http.Server
	port   int
}

func New() (*App, error) {
	cfg := config.Load()

	newsRepo := storage.NewInMemoryNewsRepository()
	rewriterClient := rewriter.NewNoopClient()
	sourceClient := buildSourceClient(cfg)

	newsUseCase := usecase.NewNewsUseCase(newsRepo, sourceClient, rewriterClient)
	handler := httptransport.NewHandler(newsUseCase)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           handler.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server: server,
		port:   cfg.HTTP.Port,
	}, nil
}

func (a *App) Run() error {
	log.Printf("starting Cash Factories backend")
	log.Printf("http address: http://localhost:%d", a.port)
	log.Printf("health check: http://localhost:%d/health", a.port)
	log.Printf("swagger docs: http://localhost:%d/docs", a.port)
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func buildSourceClient(cfg config.Config) domain.SourceClient {
	if cfg.Source.GuardianAPIKey != "" {
		log.Printf("news source: The Guardian")
		return source.NewGuardianClient(cfg.Source.GuardianAPIKey)
	}

	log.Printf("news source: stub (GUARDIAN_API_KEY is not set)")
	return source.NewStubClient()
}
