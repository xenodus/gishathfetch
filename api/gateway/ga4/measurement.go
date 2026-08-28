package ga4

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/logger"
)

const (
	// DefaultMeasurementID matches frontend/index.html gtag config.
	DefaultMeasurementID = "G-6NRLSYZ9P9"
	defaultCollectURL    = "https://www.google-analytics.com/mp/collect"
	telegramClientID     = "telegram-bot"
)

// MeasurementSender posts GA4 events via the Measurement Protocol.
type MeasurementSender struct {
	collectURL    string
	measurementID string
	apiSecret     string
	httpClient    *http.Client
}

// NewMeasurementSender builds a sender from environment configuration.
// Requires GA4_MEASUREMENT_API_SECRET; measurement ID defaults to DefaultMeasurementID.
func NewMeasurementSender() (*MeasurementSender, error) {
	apiSecret := config.GA4MeasurementAPISecret()
	if apiSecret == "" {
		return nil, fmt.Errorf("ga4: %s is not set", config.GA4MeasurementAPISecretEnv)
	}

	measurementID := config.GA4MeasurementID()
	if measurementID == "" {
		measurementID = DefaultMeasurementID
	}

	return &MeasurementSender{
		collectURL:    defaultCollectURL,
		measurementID: measurementID,
		apiSecret:     apiSecret,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}, nil
}

// SendSearchEvent records a search event with the same name and search_term
// parameter used by the frontend gtag integration.
func (s *MeasurementSender) SendSearchEvent(ctx context.Context, searchTerm string) error {
	searchTerm = strings.TrimSpace(searchTerm)
	if searchTerm == "" {
		return fmt.Errorf("ga4: search term is required")
	}

	payload := map[string]any{
		"client_id": telegramClientID,
		"events": []map[string]any{
			{
				"name": SearchEventName,
				"params": map[string]any{
					"search_term":          searchTerm,
					"engagement_time_msec": 1,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint, err := s.collectEndpoint()
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("ga4: measurement protocol status %d: %s", res.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func (s *MeasurementSender) collectEndpoint() (string, error) {
	query := url.Values{}
	query.Set("measurement_id", s.measurementID)
	query.Set("api_secret", s.apiSecret)
	return s.collectURL + "?" + query.Encode(), nil
}

// TrySendSearchEvent sends a search event when Measurement Protocol is configured.
// Failures are logged and do not propagate to callers.
func TrySendSearchEvent(ctx context.Context, searchTerm string) {
	if !config.GA4MeasurementConfigured() {
		return
	}

	sender, err := NewMeasurementSender()
	if err != nil {
		logger.From(ctx).WarnContext(ctx, "ga4 search event skipped", "err", err)
		return
	}

	if err := sender.SendSearchEvent(ctx, searchTerm); err != nil {
		logger.From(ctx).WarnContext(ctx, "ga4 search event failed", "searchTerm", searchTerm, "err", err)
		return
	}

	logger.From(ctx).InfoContext(ctx, "ga4 telegram search event sent", "searchTerm", searchTerm)
}
