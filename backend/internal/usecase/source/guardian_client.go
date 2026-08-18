package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cashfac-test/internal/domain"
)

const guardianBaseURL = "https://content.guardianapis.com/search"

type GuardianClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewGuardianClient(apiKey string) *GuardianClient {
	return &GuardianClient{
		apiKey:  apiKey,
		baseURL: guardianBaseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *GuardianClient) FetchLatest(ctx context.Context, limit int) ([]domain.SourceItem, error) {
	if limit <= 0 {
		limit = 10
	}

	requestURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse guardian url: %w", err)
	}

	query := requestURL.Query()
	query.Set("api-key", c.apiKey)
	query.Set("page-size", fmt.Sprintf("%d", limit))
	query.Set("order-by", "newest")
	query.Set("type", "article")
	query.Set("show-fields", "headline,trailText,bodyText")
	requestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build guardian request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform guardian request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("guardian api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload guardianSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode guardian response: %w", err)
	}

	items := make([]domain.SourceItem, 0, len(payload.Response.Results))
	for _, result := range payload.Response.Results {
		text := strings.TrimSpace(result.Fields.BodyText)
		if text == "" {
			text = strings.TrimSpace(result.Fields.TrailText)
		}
		if text == "" {
			continue
		}

		title := strings.TrimSpace(result.Fields.Headline)
		if title == "" {
			title = strings.TrimSpace(result.WebTitle)
		}

		items = append(items, domain.SourceItem{
			ExternalID:   result.ID,
			Title:        title,
			Text:         text,
			SourceName:   "The Guardian",
			SourceURL:    result.WebURL,
			PublishedRaw: result.WebPublicationDate,
		})
	}

	return items, nil
}

type guardianSearchResponse struct {
	Response struct {
		Results []guardianResult `json:"results"`
	} `json:"response"`
}

type guardianResult struct {
	ID                 string `json:"id"`
	WebTitle           string `json:"webTitle"`
	WebURL             string `json:"webUrl"`
	WebPublicationDate string `json:"webPublicationDate"`
	Fields             struct {
		Headline  string `json:"headline"`
		TrailText string `json:"trailText"`
		BodyText  string `json:"bodyText"`
	} `json:"fields"`
}
