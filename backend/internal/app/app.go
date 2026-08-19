package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cashfac-test/internal/config"
	"cashfac-test/internal/domain"
	"cashfac-test/internal/platform/logger"
	httptransport "cashfac-test/internal/transport/http"
	"cashfac-test/internal/usecase"
	"cashfac-test/internal/usecase/rewriter"
	"cashfac-test/internal/usecase/source"
	"cashfac-test/internal/usecase/storage"
)

type App struct {
	server       *http.Server
	port         int
	sourceName   string
	rewriterName string
	storagePath  string
}

func New() (*App, error) {
	cfg := config.Load()

	newsRepo, err := storage.NewSQLiteNewsRepository(cfg.Store.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("init sqlite repository: %w", err)
	}
	jobRepo := storage.NewInMemorySyncJobRepository()
	rewriterClient, rewriterName := buildRewriterClient(cfg)
	sourceClient, sourceName := buildSourceClient(cfg)

	newsUseCase := usecase.NewNewsUseCase(newsRepo, sourceClient, rewriterClient)
	syncJobsUseCase := usecase.NewSyncJobsUseCase(jobRepo, newsUseCase)
	handler := httptransport.NewHandler(newsUseCase, syncJobsUseCase)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           httptransport.WithRequestLogging(handler.Router()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server:       server,
		port:         cfg.HTTP.Port,
		sourceName:   sourceName,
		rewriterName: rewriterName,
		storagePath:  cfg.Store.SQLitePath,
	}, nil
}

func (a *App) Run() error {
	baseURL := fmt.Sprintf("http://localhost:%d", a.port)
	logger.Banner("Cash Factories News API",
		logger.F("server", baseURL),
		logger.F("health", baseURL+"/health"),
		logger.F("swagger", baseURL+"/docs"),
		logger.F("source", a.sourceName),
		logger.F("rewriter", a.rewriterName),
		logger.F("storage", "SQLite · "+a.storagePath),
	)
	logger.Success("BOOT", "server is ready", logger.F("port", a.port))
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}

func buildSourceClient(cfg config.Config) (domain.SourceClient, string) {
	if cfg.Source.GuardianAPIKey != "" {
		return source.NewGuardianClient(cfg.Source.GuardianAPIKey), "The Guardian"
	}

	logger.Warn("CONFIG", "Guardian API key is missing; using stub source")
	return source.NewStubClient(), "Stub"
}

func buildRewriterClient(cfg config.Config) (domain.Rewriter, string) {
	if cfg.AI.ZAIAPIKey != "" {
		return rewriter.NewZAIClient(cfg.AI.ZAIAPIKey, cfg.AI.ZAIBaseURL, cfg.AI.ZAIModel), "Z.ai · " + cfg.AI.ZAIModel
	}

	logger.Warn("CONFIG", "Z.ai API key is missing; rewrites are disabled")
	return rewriter.NewNoopClient(), "Disabled"
}
