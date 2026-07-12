package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func TestSendDiscordAlert(t *testing.T) {
	testSendDiscordWebhook(t, DiscordWebhookURLEnv, SendDiscordAlert)
}

func TestSendJobDiscordAlert(t *testing.T) {
	testSendDiscordWebhook(t, JobDiscordWebhookURLEnv, SendJobDiscordAlert)
}

func testSendDiscordWebhook(t *testing.T, webhookURLEnv string, sendAlert func(string)) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			var payload DiscordPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if payload.Content != "Test Message" {
				t.Errorf("Expected content 'Test Message', got '%s'", payload.Content)
			}

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		t.Setenv(webhookURLEnv, server.URL)
		sendAlert("Test Message")
	})

	t.Run("No URL Set", func(t *testing.T) {
		t.Setenv(webhookURLEnv, "")
		sendAlert("Should result in log but no panic")
	})

	t.Run("Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		t.Setenv(webhookURLEnv, server.URL)
		sendAlert("Test Message")
	})
}

func TestSendDiscordAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv(DiscordWebhookURLEnv)
	if webhookURL == "" {
		t.Skip("DISCORD_WEBHOOK_URL not set, skipping integration test")
	}

	SendDiscordAlert("Integration Test Message (Ignore this)")
}

func TestSendJobDiscordAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv(JobDiscordWebhookURLEnv)
	if webhookURL == "" {
		t.Skip("JOB_DISCORD_WEBHOOK_URL not set, skipping integration test")
	}

	SendJobDiscordAlert("Integration Test Message (Ignore this)")
}
