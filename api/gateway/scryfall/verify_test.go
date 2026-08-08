package scryfall

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestVerifyCardNameAccentInsensitiveAutocomplete(t *testing.T) {
	origHTTPGet := httpGet
	t.Cleanup(func() { httpGet = origHTTPGet })

	httpGet = func(_ context.Context, requestURL string) (*http.Response, error) {
		if strings.Contains(requestURL, "/cards/autocomplete") {
			return jsonResponse(http.StatusOK, `{"object":"catalog","data":["Kíli the Resourceful"]}`), nil
		}
		t.Fatalf("unexpected request URL: %s", requestURL)
		return nil, nil
	}

	got, err := VerifyCardName(context.Background(), "Kili the Resourceful")
	if err != nil {
		t.Fatalf("VerifyCardName() error = %v", err)
	}
	if got != "Kíli the Resourceful" {
		t.Fatalf("VerifyCardName() = %q, want %q", got, "Kíli the Resourceful")
	}
}

func TestVerifyCardNameFuzzyFallback(t *testing.T) {
	origHTTPGet := httpGet
	t.Cleanup(func() { httpGet = origHTTPGet })

	httpGet = func(_ context.Context, requestURL string) (*http.Response, error) {
		switch {
		case strings.Contains(requestURL, "/cards/autocomplete"):
			return jsonResponse(http.StatusOK, `{"object":"catalog","data":[]}`), nil
		case strings.Contains(requestURL, "/cards/named?exact="):
			return jsonResponse(http.StatusNotFound, `{"object":"error","code":"not_found"}`), nil
		case strings.Contains(requestURL, "/cards/named?fuzzy="):
			return jsonResponse(http.StatusOK, `{"object":"card","name":"Juzám Djinn"}`), nil
		default:
			t.Fatalf("unexpected request URL: %s", requestURL)
			return nil, nil
		}
	}

	got, err := VerifyCardName(context.Background(), "Juzam Djinn")
	if err != nil {
		t.Fatalf("VerifyCardName() error = %v", err)
	}
	if got != "Juzám Djinn" {
		t.Fatalf("VerifyCardName() = %q, want %q", got, "Juzám Djinn")
	}
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
