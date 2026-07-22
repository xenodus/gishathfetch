package alert

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const (
	SlackAlertWebhookEnv    = "SLACK_ALERT_WEBHOOK"
	SlackJobWebhookURLEnv   = "SLACK_JOB_WEBHOOK"
)

type SlackPayload struct {
	Text string `json:"text"`
}

// SendSlackAlert sends a message to the search-error Slack webhook.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendSlackAlert(message string) {
	sendSlackWebhook(SlackAlertWebhookEnv, message)
}

// SendJobSlackAlert sends a message to the scheduled-job Slack webhook.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendJobSlackAlert(message string) {
	sendSlackWebhook(SlackJobWebhookURLEnv, message)
}

func sendSlackWebhook(webhookURLEnv, message string) {
	webhookURL := os.Getenv(webhookURLEnv)
	if webhookURL == "" {
		log.Printf("%s not set, skipping alert", webhookURLEnv)
		return
	}

	payload := SlackPayload{
		Text: message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal slack payload: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Failed to send slack alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Slack API returned non-2xx status: %d", resp.StatusCode)
	}
}
