package alert

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

const (
	DiscordWebhookURLEnv    = "DISCORD_WEBHOOK_URL"
	JobDiscordWebhookURLEnv = "JOB_DISCORD_WEBHOOK_URL"
)

type DiscordPayload struct {
	Content string `json:"content"`
}

// SendDiscordAlert sends a message to the search-error Discord webhook.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendDiscordAlert(message string) {
	sendDiscordWebhook(DiscordWebhookURLEnv, message)
}

// SendJobDiscordAlert sends a message to the scheduled-job Discord webhook.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendJobDiscordAlert(message string) {
	sendDiscordWebhook(JobDiscordWebhookURLEnv, message)
}

func sendDiscordWebhook(webhookURLEnv, message string) {
	webhookURL := os.Getenv(webhookURLEnv)
	if webhookURL == "" {
		log.Printf("%s not set, skipping alert", webhookURLEnv)
		return
	}

	payload := DiscordPayload{
		Content: message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal discord payload: %v", err)
		return
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Failed to send discord alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Discord API returned non-200 status: %d", resp.StatusCode)
	}
}
