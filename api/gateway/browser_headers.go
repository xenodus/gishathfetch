package gateway

import (
	"net/http"
	"net/url"
)

const (
	browserLikeAcceptHTML     = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"
	browserLikeAcceptJSON     = "application/json, text/javascript, */*;q=0.01"
	browserLikeAcceptLanguage = "en-US,en;q=0.9"
)

// navigationReferer returns scheme://host/ for use as Referer when emulating same-origin navigation.
func navigationReferer(target *url.URL) string {
	if target == nil || target.Scheme == "" || target.Host == "" {
		return ""
	}
	return target.Scheme + "://" + target.Host + "/"
}

// ApplyBrowserLikeHTMLHeaders sets headers typical of a top-level browser document request.
func ApplyBrowserLikeHTMLHeaders(h *http.Header, pageURL *url.URL) {
	if h == nil {
		return
	}
	if ref := navigationReferer(pageURL); ref != "" {
		h.Set("Referer", ref)
	}
	h.Set("Accept", browserLikeAcceptHTML)
	h.Set("Accept-Language", browserLikeAcceptLanguage)
	h.Set("Upgrade-Insecure-Requests", "1")
}

// ApplyBrowserLikeJSONFetchHeaders sets headers typical of in-page JSON/XHR requests to a store domain.
// storeBase is the public shop URL (e.g. https://example.com); it may be nil to omit Referer only.
func ApplyBrowserLikeJSONFetchHeaders(h *http.Header, storeBase *url.URL) {
	if h == nil {
		return
	}
	if ref := navigationReferer(storeBase); ref != "" {
		h.Set("Referer", ref)
		if storeBase.Scheme != "" && storeBase.Host != "" {
			h.Set("Origin", storeBase.Scheme+"://"+storeBase.Host)
		}
	}
	h.Set("Accept", browserLikeAcceptJSON)
	h.Set("Accept-Language", browserLikeAcceptLanguage)
}
