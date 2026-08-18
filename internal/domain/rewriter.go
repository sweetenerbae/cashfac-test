package domain

import "context"

type RewriteRequest struct {
	Title string
	Text  string
	Mood  Mood
}

type RewriteResponse struct {
	Text         string
	FactChecksum string
}

type Rewriter interface {
	Rewrite(ctx context.Context, request RewriteRequest) (RewriteResponse, error)
}
