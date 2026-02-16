package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
		os.Setenv("DISCORD_WEBHOOK_URL", server.URL)
		defer os.Unsetenv("DISCORD_WEBHOOK_URL")

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
