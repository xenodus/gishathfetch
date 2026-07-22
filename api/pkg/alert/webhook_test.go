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

func TestSendAlert(t *testing.T) {
	testSendWebhookAlert(t, AlertWebhookEnv, SendAlert)
}

func TestSendJobAlert(t *testing.T) {
	testSendWebhookAlert(t, JobAlertWebhookEnv, SendJobAlert)
}

func testSendWebhookAlert(t *testing.T, webhookURLEnv string, sendAlert func(string)) {
	t.Helper()

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			var payload WebhookPayload
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

func TestSendAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv(AlertWebhookEnv)
	if webhookURL == "" {
		t.Skip("SLACK_ALERT_WEBHOOK not set, skipping integration test")
	}

	SendAlert("Integration Test Message (Ignore this)")
}

func TestSendJobAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv(JobAlertWebhookEnv)
	if webhookURL == "" {
		t.Skip("SLACK_JOB_WEBHOOK not set, skipping integration test")
	}

	SendJobAlert("Integration Test Message (Ignore this)")
}
