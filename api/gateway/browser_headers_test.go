package gateway

import (
	"net/http"
	"net/url"
	"testing"
)

func TestApplyBrowserLikeHTMLHeaders(t *testing.T) {
	u, err := url.Parse("https://shop.example/path?q=1")
	if err != nil {
		t.Fatal(err)
	}
	h := make(http.Header)
	ApplyBrowserLikeHTMLHeaders(&h, u)

	if got := h.Get("Referer"); got != "https://shop.example/" {
		t.Fatalf("Referer: got %q", got)
	}
	if got := h.Get("Accept"); got != browserLikeAcceptHTML {
		t.Fatalf("Accept: got %q", got)
	}
	if got := h.Get("Accept-Language"); got != browserLikeAcceptLanguage {
		t.Fatalf("Accept-Language: got %q", got)
	}
	if got := h.Get("Upgrade-Insecure-Requests"); got != "1" {
		t.Fatalf("Upgrade-Insecure-Requests: got %q", got)
	}
}

func TestApplyBrowserLikeHTMLHeaders_nilHeader(t *testing.T) {
	u, _ := url.Parse("https://a.test/")
	ApplyBrowserLikeHTMLHeaders(nil, u) // must not panic
}

func TestApplyBrowserLikeJSONFetchHeaders(t *testing.T) {
	base, err := url.Parse("https://store.example")
	if err != nil {
		t.Fatal(err)
	}
	h := make(http.Header)
	ApplyBrowserLikeJSONFetchHeaders(&h, base)

	if got := h.Get("Referer"); got != "https://store.example/" {
		t.Fatalf("Referer: got %q", got)
	}
	if got := h.Get("Accept"); got != browserLikeAcceptJSON {
		t.Fatalf("Accept: got %q", got)
	}
	if got := h.Get("Accept-Encoding"); got != "gzip" {
		t.Fatalf("Accept-Encoding: got %q", got)
	}
}

func TestApplyBrowserLikeJSONFetchHeaders_nilStoreOmitsReferer(t *testing.T) {
	h := make(http.Header)
	ApplyBrowserLikeJSONFetchHeaders(&h, nil)
	if h.Get("Referer") != "" {
		t.Fatalf("expected empty Referer, got %q", h.Get("Referer"))
	}
	if h.Get("Accept-Language") == "" {
		t.Fatal("expected Accept-Language")
	}
}
