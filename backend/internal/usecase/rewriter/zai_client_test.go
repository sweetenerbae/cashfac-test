package rewriter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"cashfac-test/internal/domain"
)

func TestZAIClientDoesNotRetryWeakRewrite(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"Original story text."}}]}`)
	}))
	defer server.Close()

	client := NewZAIClient("test-key", server.URL, "test-model")
	_, err := client.Rewrite(context.Background(), domain.RewriteRequest{
		Title: "Sample title",
		Text:  "Original story text.",
		Mood:  domain.MoodNeutral,
	})
	if err == nil {
		t.Fatal("expected weak rewrite error")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly one Z.ai request, got %d", got)
	}
}
