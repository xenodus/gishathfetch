package alert

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

const (
	AlertWebhookEnv    = "SLACK_ALERT_WEBHOOK"
	JobAlertWebhookEnv = "SLACK_JOB_WEBHOOK"
)

type WebhookPayload struct {
	Text string `json:"text"`
}

// SendAlert sends a message to the search-error alert webhook.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendAlert(message string) {
	sendWebhookAlert(AlertWebhookEnv, message)
}

// SendJobAlert sends a message to the scheduled-job alert webhook.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendJobAlert(message string) {
	sendWebhookAlert(JobAlertWebhookEnv, message)
}

func sendWebhookAlert(webhookURLEnv, message string) {
	webhookURL := os.Getenv(webhookURLEnv)
	if webhookURL == "" {
		slog.Warn("alert webhook not set, skipping alert", "env", webhookURLEnv)
		return
	}

	payload := WebhookPayload{
		Text: message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal alert payload", "err", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error("failed to send alert", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("alert webhook returned non-2xx status", "status", resp.StatusCode)
	}
}
