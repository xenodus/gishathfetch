package alert

import (
	"bytes"
	"encoding/json"
	"log"
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

// sendAlertAsync posts to the search-error webhook without blocking the caller.
// Tests may replace this to observe notifications synchronously.
var sendAlertAsync = func(message string) {
	go SendAlert(message)
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

// NotifySearchStrategyFallback logs a search-strategy fallback and fire-and-forget
// alerts via SLACK_ALERT_WEBHOOK so mid-chain recoveries are visible in Slack
// even when the store search eventually succeeds.
func NotifySearchStrategyFallback(message string) {
	log.Println(message)
	sendAlertAsync(message)
}

func sendWebhookAlert(webhookURLEnv, message string) {
	webhookURL := os.Getenv(webhookURLEnv)
	if webhookURL == "" {
		log.Printf("%s not set, skipping alert", webhookURLEnv)
		return
	}

	payload := WebhookPayload{
		Text: message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal alert payload: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Failed to send alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Alert webhook returned non-2xx status: %d", resp.StatusCode)
	}
}
