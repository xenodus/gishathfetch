package telegrambot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SearchSummary is the Telegram search API response payload.
type SearchSummary struct {
	Cheapest        *CardSummary `json:"cheapest"`
	ResultCount     int          `json:"resultCount"`
	WebsiteURL      string       `json:"websiteUrl"`
	TotalDurationMs int64        `json:"totalDurationMs"`
}

// CardSummary is the cheapest card returned by /telegram/search.
type CardSummary struct {
	Name      string  `json:"name"`
	URL       string  `json:"url"`
	Img       string  `json:"img"`
	Price     float64 `json:"price"`
	InStock   bool    `json:"inStock"`
	IsFoil    bool    `json:"isFoil"`
	Source    string  `json:"src"`
	Quality   string  `json:"quality"`
	ExtraInfo string  `json:"extraInfo"`
}

// GishathClient calls the Telegram search endpoint.
type GishathClient struct {
	baseURL      string
	botToken     string
	originSecret string
	httpClient   *http.Client
}

// NewGishathClient builds a client for /telegram/search.
func NewGishathClient(baseURL, botToken, originSecret string) *GishathClient {
	return &GishathClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		botToken:     botToken,
		originSecret: originSecret,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

// Search runs a card search and returns the summary payload.
func (c *GishathClient) Search(ctx context.Context, query string) (*SearchSummary, error) {
	endpoint := c.baseURL + "/telegram/search?" + url.Values{"s": {query}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.botToken)
	if c.originSecret != "" {
		req.Header.Set("X-Origin-Verify", c.originSecret)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gishath search failed: status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var summary SearchSummary
	if err := json.Unmarshal(body, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}
