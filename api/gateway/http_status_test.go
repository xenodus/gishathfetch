package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestFormatHTTPStatus(t *testing.T) {
	if got := FormatHTTPStatus(http.StatusForbidden); got != "403 Forbidden" {
		t.Fatalf("FormatHTTPStatus(403) = %q, want %q", got, "403 Forbidden")
	}

	if got := FormatHTTPStatus(150); got != "status=150" {
		t.Fatalf("FormatHTTPStatus(150) = %q, want %q", got, "status=150")
	}
}

func TestExtractHTTPStatusCode(t *testing.T) {
	tests := map[string]int{
		"unexpected status 429":                                 429,
		"binderpos decklist request failed status=503 body=...": 503,
		"unexpected status for Cards Central: 503 Service Unavailable": 503,
		"403 Forbidden (proxy_mode=direct proxy=none)":            403,
		"Service Unavailable":                                     0,
	}

	for msg, want := range tests {
		if got := ExtractHTTPStatusCode(msg); got != want {
			t.Fatalf("ExtractHTTPStatusCode(%q) = %d, want %d", msg, got, want)
		}
	}
}

func TestIsHTTPServerError(t *testing.T) {
	tests := map[string]bool{
		"503 Service Unavailable":                                                      true,
		"attempt 1 (scrap-dedicated): 503 Service Unavailable (proxy_mode=direct)":     true,
		"attempt 2 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)": true,
		"403 Forbidden":           false,
		"unexpected status 429":   false,
		"connection reset by peer": false,
	}

	for msg, want := range tests {
		if got := IsHTTPServerError(errors.New(msg)); got != want {
			t.Fatalf("IsHTTPServerError(%q) = %v, want %v", msg, got, want)
		}
	}

	wrapped := fmt.Errorf("attempt 1 (scrap-dedicated): %w", errors.New("503 Service Unavailable"))
	if !IsHTTPServerError(wrapped) {
		t.Fatalf("IsHTTPServerError() should detect wrapped 5xx errors")
	}
}

func TestEnrichErrorWithHTTPStatus(t *testing.T) {
	err := EnrichErrorWithHTTPStatus(errors.New(http.StatusText(http.StatusForbidden)), http.StatusForbidden)
	if got := err.Error(); got != "403 Forbidden" {
		t.Fatalf("EnrichErrorWithHTTPStatus() = %q, want %q", got, "403 Forbidden")
	}

	alreadyHasCode := EnrichErrorWithHTTPStatus(
		errors.New("unexpected status 429"),
		http.StatusTooManyRequests,
	)
	if got := alreadyHasCode.Error(); got != "unexpected status 429" {
		t.Fatalf("expected existing status to remain unchanged, got %q", got)
	}
}

func TestEnsureHTTPStatusInErrorMessage(t *testing.T) {
	if got := EnsureHTTPStatusInErrorMessage("Forbidden"); got != "403 Forbidden" {
		t.Fatalf("EnsureHTTPStatusInErrorMessage(Forbidden) = %q", got)
	}

	if got := EnsureHTTPStatusInErrorMessage("attempt 2 (scrap-direct): Service Unavailable (proxy_mode=direct proxy=none)"); got != "attempt 2 (scrap-direct): 503 Service Unavailable (proxy_mode=direct proxy=none)" {
		t.Fatalf("EnsureHTTPStatusInErrorMessage() = %q", got)
	}

	if got := EnsureHTTPStatusInErrorMessage("403 Forbidden (proxy_mode=direct proxy=none)"); got != "403 Forbidden (proxy_mode=direct proxy=none)" {
		t.Fatalf("EnsureHTTPStatusInErrorMessage() should not double-prefix, got %q", got)
	}
}
