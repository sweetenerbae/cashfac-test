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

type RewriteResultSource string

const (
	RewriteSourceGenerated RewriteResultSource = "generated"
	RewriteSourceCache     RewriteResultSource = "cache"
	RewriteSourceShared    RewriteResultSource = "shared"
)

type RewriteMeta struct {
	Source           RewriteResultSource
	LLMRequests      int
	SavedLLMRequests int
	DurationMs       int64
}

type RewriteResult struct {
	News News
	Meta RewriteMeta
}

type Rewriter interface {
	Rewrite(ctx context.Context, request RewriteRequest) (RewriteResponse, error)
}
