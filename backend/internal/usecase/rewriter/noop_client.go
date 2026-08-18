package rewriter

import (
	"context"
	"fmt"

	"cashfac-test/internal/domain"
)

type NoopClient struct{}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (c *NoopClient) Rewrite(_ context.Context, request domain.RewriteRequest) (domain.RewriteResponse, error) {
	text := fmt.Sprintf("[%s] %s", request.Mood, request.Text)

	return domain.RewriteResponse{
		Text:         text,
		FactChecksum: request.Title + ":" + request.Text,
	}, nil
}
