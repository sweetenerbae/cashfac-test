package http

import (
	standardhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestLoggingCapturesResponse(t *testing.T) {
	handler := WithRequestLogging(standardhttp.HandlerFunc(func(w standardhttp.ResponseWriter, _ *standardhttp.Request) {
		w.WriteHeader(standardhttp.StatusCreated)
		_, _ = w.Write([]byte("created"))
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(standardhttp.MethodPost, "/api/v1/news/sync?limit=10", nil))

	if response.Code != standardhttp.StatusCreated {
		t.Fatalf("expected status %d, got %d", standardhttp.StatusCreated, response.Code)
	}
	if requestID := response.Header().Get("X-Request-ID"); !strings.HasPrefix(requestID, "req-") {
		t.Fatalf("expected generated request id, got %q", requestID)
	}
}

func TestWithRequestLoggingRecoversFromPanic(t *testing.T) {
	handler := WithRequestLogging(standardhttp.HandlerFunc(func(standardhttp.ResponseWriter, *standardhttp.Request) {
		panic("unexpected failure")
	}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, httptest.NewRequest(standardhttp.MethodGet, "/panic", nil))

	if response.Code != standardhttp.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", standardhttp.StatusInternalServerError, response.Code)
	}
}
