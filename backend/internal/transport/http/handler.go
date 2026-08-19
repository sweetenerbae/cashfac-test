package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/usecase"
)

type Handler struct {
	newsUseCase     *usecase.NewsUseCase
	syncJobsUseCase *usecase.SyncJobsUseCase
	mux             *http.ServeMux
}

func NewHandler(newsUseCase *usecase.NewsUseCase, syncJobsUseCase *usecase.SyncJobsUseCase) *Handler {
	h := &Handler{
		newsUseCase:     newsUseCase,
		syncJobsUseCase: syncJobsUseCase,
		mux:             http.NewServeMux(),
	}

	h.registerRoutes()

	return h
}

func (h *Handler) Router() http.Handler {
	return h.mux
}

func (h *Handler) registerRoutes() {
	h.mux.HandleFunc("/health", h.handleHealth)
	h.mux.HandleFunc("/docs", h.handleSwaggerUI)
	h.mux.HandleFunc("/docs/", h.handleSwaggerUI)
	h.mux.HandleFunc("/docs/swagger-ui.css", h.handleSwaggerCSS)
	h.mux.HandleFunc("/docs/swagger-ui-bundle.js", h.handleSwaggerJS)
	h.mux.HandleFunc("/openapi.yaml", h.handleOpenAPISpec)
	h.mux.HandleFunc("/api/v1/news/by-external", h.handleGetNewsByExternalID)
	h.mux.HandleFunc("/api/v1/news/rewrite", h.handleRewriteNews)
	h.mux.HandleFunc("/api/v1/news/sync", h.handleSyncNews)
	h.mux.HandleFunc("/api/v1/jobs/", h.handleGetJob)
	h.mux.HandleFunc("/api/v1/news/", h.handleGetNews)
	h.mux.HandleFunc("/api/v1/news", h.handleListNews)
}

func (h *Handler) handleGetNewsByExternalID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	externalID := strings.TrimSpace(r.URL.Query().Get("external_id"))
	if externalID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "external_id is required"})
		return
	}

	mood := domain.Mood(r.URL.Query().Get("mood"))

	item, err := h.newsUseCase.GetByExternalID(r.Context(), externalID, mood)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNewsNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) handleRewriteNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	externalID := strings.TrimSpace(r.URL.Query().Get("external_id"))
	if externalID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "external_id is required"})
		return
	}

	mood := domain.Mood(r.URL.Query().Get("mood"))
	if mood == "" {
		mood = domain.MoodNeutral
	}
	if !mood.IsValid() {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported mood"})
		return
	}

	result, err := h.newsUseCase.RewriteByExternalID(r.Context(), externalID, mood)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNewsNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/docs" && r.URL.Path != "/docs/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(swaggerHTML)
}

func (h *Handler) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/openapi.yaml" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func (h *Handler) handleSwaggerCSS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/docs/swagger-ui.css" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(swaggerCSS)
}

func (h *Handler) handleSwaggerJS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/docs/swagger-ui-bundle.js" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(swaggerJS)
}

func (h *Handler) handleSyncNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	limit := 10
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be a positive integer"})
			return
		}
		limit = parsedLimit
	}

	job, err := h.syncJobsUseCase.Start(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, job)
}

func (h *Handler) handleListNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	mood := domain.Mood(r.URL.Query().Get("mood"))

	items, err := h.newsUseCase.List(r.Context(), mood)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) handleGetNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/news/")
	if id == "" || id == "sync" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "news not found"})
		return
	}

	item, err := h.newsUseCase.Get(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrNewsNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) handleGetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/v1/jobs/")
	if id == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}

	job, err := h.syncJobsUseCase.Get(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, domain.ErrJobNotFound) {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
