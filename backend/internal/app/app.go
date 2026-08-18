package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
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
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	cfg := config.Load()

	newsRepo, err := storage.NewSQLiteNewsRepository(cfg.Store.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("init sqlite repository: %w", err)
	}
	jobRepo := storage.NewInMemorySyncJobRepository()
	rewriterClient := buildRewriterClient(cfg)
	sourceClient := buildSourceClient(cfg)

	newsUseCase := usecase.NewNewsUseCase(newsRepo, sourceClient, rewriterClient)
	syncJobsUseCase := usecase.NewSyncJobsUseCase(jobRepo, newsUseCase)
	handler := httptransport.NewHandler(newsUseCase, syncJobsUseCase)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           withRequestLogging(handler.Router()),
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
	log.Printf("storage: sqlite")
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

func buildRewriterClient(cfg config.Config) domain.Rewriter {
	if cfg.AI.ZAIAPIKey != "" {
		log.Printf("rewriter: Z.ai (%s)", cfg.AI.ZAIModel)
		return rewriter.NewZAIClient(cfg.AI.ZAIAPIKey, cfg.AI.ZAIBaseURL, cfg.AI.ZAIModel)
	}

	log.Printf("rewriter: noop (ZAI_API_KEY is not set)")
	return rewriter.NewNoopClient()
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("request: method=%s path=%s duration=%s", r.Method, r.URL.RequestURI(), time.Since(startedAt).Round(time.Millisecond))
	})
}
