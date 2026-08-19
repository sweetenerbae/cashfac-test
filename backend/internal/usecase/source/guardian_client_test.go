package source

import "testing"

func TestBestImageURLSelectsClosestAssetAtOrAboveTarget(t *testing.T) {
	result := guardianResult{
		Elements: []guardianElement{
			{
				Type: "image",
				Assets: []guardianAsset{
					{File: "https://example.com/500.jpg", TypeData: map[string]any{"width": float64(500)}},
					{File: "https://example.com/1600.jpg", TypeData: map[string]any{"width": "1600"}},
					{File: "https://example.com/1200.jpg", TypeData: map[string]any{"width": float64(1200)}},
				},
			},
		},
	}

	if got := bestImageURL(result); got != "https://example.com/1200.jpg" {
		t.Fatalf("expected 1200px asset, got %q", got)
	}
}

func TestBestImageURLFallsBackToThumbnail(t *testing.T) {
	result := guardianResult{}
	result.Fields.Thumbnail = " https://example.com/thumbnail.jpg "

	if got := bestImageURL(result); got != "https://example.com/thumbnail.jpg" {
		t.Fatalf("expected thumbnail fallback, got %q", got)
	}
}
