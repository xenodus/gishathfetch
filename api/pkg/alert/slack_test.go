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

func TestSendSlackAlert(t *testing.T) {
	testSendSlackWebhook(t, SlackAlertWebhookEnv, SendSlackAlert)
}

func TestSendJobSlackAlert(t *testing.T) {
	testSendSlackWebhook(t, SlackJobWebhookURLEnv, SendJobSlackAlert)
}

func testSendSlackWebhook(t *testing.T, webhookURLEnv string, sendAlert func(string)) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			var payload SlackPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode request body: %v", err)
			}

			if payload.Text != "Test Message" {
				t.Errorf("Expected text 'Test Message', got '%s'", payload.Text)
			}

			w.WriteHeader(http.StatusOK)
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

func TestSendSlackAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv(SlackAlertWebhookEnv)
	if webhookURL == "" {
		t.Skip("SLACK_ALERT_WEBHOOK not set, skipping integration test")
	}

	SendSlackAlert("Integration Test Message (Ignore this)")
}

func TestSendJobSlackAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv(SlackJobWebhookURLEnv)
	if webhookURL == "" {
		t.Skip("SLACK_JOB_WEBHOOK not set, skipping integration test")
	}

	SendJobSlackAlert("Integration Test Message (Ignore this)")
}
