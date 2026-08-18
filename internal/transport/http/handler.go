package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"cashfac-test/internal/domain"
	"cashfac-test/internal/usecase"
)

type Handler struct {
	newsUseCase *usecase.NewsUseCase
	mux         *http.ServeMux
}

func NewHandler(newsUseCase *usecase.NewsUseCase) *Handler {
	h := &Handler{
		newsUseCase: newsUseCase,
		mux:         http.NewServeMux(),
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
	h.mux.HandleFunc("/openapi.yaml", h.handleOpenAPISpec)
	h.mux.HandleFunc("/api/v1/news/sync", h.handleSyncNews)
	h.mux.HandleFunc("/api/v1/news/", h.handleGetNews)
	h.mux.HandleFunc("/api/v1/news", h.handleListNews)
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

func (h *Handler) handleSyncNews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	mood := domain.Mood(r.URL.Query().Get("mood"))
	if mood == "" {
		mood = domain.MoodNeutral
	}

	if err := h.newsUseCase.Sync(r.Context(), 10, mood); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "sync started"})
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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
