package gateway

import (
	"net/http"
	"strings"
)

const (
	shopifyCartCurrencyCookie   = "cart_currency=SGD"
	shopifyLocalizationCookie = "localization=SG"
)

// ApplyShopifySGDCurrencyCookie sets Shopify storefront cookies that force SGD pricing.
func ApplyShopifySGDCurrencyCookie(h *http.Header) {
	if h == nil {
		return
	}
	appendCookieHeader(h, shopifyCartCurrencyCookie)
	appendCookieHeader(h, shopifyLocalizationCookie)
}

func appendCookieHeader(h *http.Header, cookie string) {
	cookie = strings.TrimSpace(cookie)
	if cookie == "" {
		return
	}

	existing := strings.TrimSpace(h.Get("Cookie"))
	if existing == "" {
		h.Set("Cookie", cookie)
		return
	}
	h.Set("Cookie", existing+"; "+cookie)
}
