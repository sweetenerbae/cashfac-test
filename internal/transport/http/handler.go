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
	_, _ = w.Write([]byte(swaggerHTML))
}

func (h *Handler) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/openapi.yaml" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(openAPISpec))
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

const swaggerHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Cashfac News API Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui"
      });
    </script>
  </body>
</html>`

const openAPISpec = `openapi: 3.0.3
info:
  title: Cashfac News API
  version: 0.1.0
  description: >
    API for collecting real news and exposing rewritten versions in different moods
    without changing facts.
servers:
  - url: http://localhost:8080
    description: Local development
paths:
  /health:
    get:
      summary: Health check
      tags:
        - System
      responses:
        "200":
          description: Service is healthy
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/HealthResponse"
  /api/v1/news:
    get:
      summary: List news
      tags:
        - News
      parameters:
        - in: query
          name: mood
          required: false
          schema:
            $ref: "#/components/schemas/Mood"
          description: Filter by rewrite mood
      responses:
        "200":
          description: News list
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/News"
        "500":
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
  /api/v1/news/{id}:
    get:
      summary: Get single news item
      tags:
        - News
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: News item
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/News"
        "404":
          description: News not found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
        "500":
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
  /api/v1/news/sync:
    post:
      summary: Sync latest news and rewrite them in selected mood
      tags:
        - News
      parameters:
        - in: query
          name: mood
          required: false
          schema:
            $ref: "#/components/schemas/Mood"
          description: Mood used for rewriting
      responses:
        "202":
          description: Sync accepted
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/StatusResponse"
        "500":
          description: Internal server error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ErrorResponse"
components:
  schemas:
    Mood:
      type: string
      enum:
        - neutral
        - happy
        - sad
        - ironic
    HealthResponse:
      type: object
      properties:
        status:
          type: string
          example: ok
      required:
        - status
    StatusResponse:
      type: object
      properties:
        status:
          type: string
          example: sync started
      required:
        - status
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
      required:
        - error
    News:
      type: object
      properties:
        ID:
          type: string
        Title:
          type: string
        OriginalText:
          type: string
        RewrittenText:
          type: string
        Mood:
          $ref: "#/components/schemas/Mood"
        SourceName:
          type: string
        SourceURL:
          type: string
          format: uri
        PublishedAt:
          type: string
          format: date-time
        CreatedAt:
          type: string
          format: date-time
        ExternalID:
          type: string
        FactChecksum:
          type: string
        OriginalDigest:
          type: string
`
