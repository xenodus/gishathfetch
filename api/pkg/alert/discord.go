package alert

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

type DiscordPayload struct {
	Content string `json:"content"`
}

// SendDiscordAlert sends a message to the configured Discord Webhook URL.
// It is fire-and-forget; errors are logged but not returned to disrupt the main flow.
func SendDiscordAlert(message string) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		// Log warning only once or just ignore if not configured?
		// Better to log so user knows why alerts aren't sending.
		log.Println("DISCORD_WEBHOOK_URL not set, skipping alert")
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
