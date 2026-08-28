package telegramphoto

import (
	"strings"
)

// Normalize returns an absolute http(s) URL when the scraper value is usable.
func Normalize(raw string) string {
	img := strings.TrimSpace(raw)
	if img == "" {
		return ""
	}
	if strings.HasPrefix(img, "//") {
		return "https:" + img
	}
	return img
}

// IsSendable reports whether Telegram sendPhoto is likely to accept the URL.
// Telegram fetches the URL server-side and rejects HTML/SVG and other non-raster types.
func IsSendable(photoURL string) bool {
	photoURL = strings.TrimSpace(photoURL)
	if photoURL == "" {
		return false
	}
	if !strings.HasPrefix(photoURL, "http://") && !strings.HasPrefix(photoURL, "https://") {
		return false
	}
	lower := strings.ToLower(photoURL)
	if strings.Contains(lower, "placehold.co/") {
		// Web search uses placehold.co fallbacks; Telegram sendPhoto rejects their SVG responses.
		return false
	}
	if strings.HasSuffix(lower, ".svg") {
		return false
	}
	return true
}

// Select returns the first sendable photo URL from the provided candidates.
func Select(candidates ...string) string {
	for _, candidate := range candidates {
		if url := Normalize(candidate); IsSendable(url) {
			return url
		}
	}
	return ""
}
