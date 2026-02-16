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
	// 1. Test Case: Success
	t.Run("Success", func(t *testing.T) {
		// Mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// specific assertions
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

			w.WriteHeader(http.StatusNoContent) // 204 is typical for webhooks
		}))
		defer server.Close()

		// Set ENV to mock server
		t.Setenv("DISCORD_WEBHOOK_URL", server.URL)

		// Call function
		SendDiscordAlert("Test Message")
	})

	// 2. Test Case: No URL set (Should not panic)
	t.Run("No URL Set", func(t *testing.T) {
		t.Setenv("DISCORD_WEBHOOK_URL", "")
		SendDiscordAlert("Should result in log but no panic")
	})

	// 3. Test Case: Server Error (Should handle gracefully)
	t.Run("Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		t.Setenv("DISCORD_WEBHOOK_URL", server.URL)

		SendDiscordAlert("Test Message")
	})
}

func TestSendDiscordAlert_Integration(t *testing.T) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		t.Skip("DISCORD_WEBHOOK_URL not set, skipping integration test")
	}

	SendDiscordAlert("Integration Test Message (Ignore this)")
}
