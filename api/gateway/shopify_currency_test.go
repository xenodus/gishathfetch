package gateway

import (
	"net/http"
	"testing"
)

func TestApplyShopifySGDCurrencyCookie(t *testing.T) {
	h := make(http.Header)
	ApplyShopifySGDCurrencyCookie(&h)

	got := h.Get("Cookie")
	if got != "cart_currency=SGD; localization=SG" {
		t.Fatalf("Cookie: got %q", got)
	}
}

func TestApplyShopifySGDCurrencyCookie_mergesExistingCookies(t *testing.T) {
	h := make(http.Header)
	h.Set("Cookie", "session=abc")
	ApplyShopifySGDCurrencyCookie(&h)

	got := h.Get("Cookie")
	if got != "session=abc; cart_currency=SGD; localization=SG" {
		t.Fatalf("Cookie: got %q", got)
	}
}

func TestApplyShopifySGDCurrencyCookie_nilHeader(t *testing.T) {
	ApplyShopifySGDCurrencyCookie(nil) // must not panic
}
