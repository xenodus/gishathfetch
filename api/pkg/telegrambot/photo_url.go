package telegrambot

import (
	"strings"
)

// normalizePhotoURL returns an absolute http(s) URL when the scraper value is usable.
func normalizePhotoURL(raw string) string {
	img := strings.TrimSpace(raw)
	if img == "" {
		return ""
	}
	if strings.HasPrefix(img, "//") {
		return "https:" + img
	}
	return img
}

// isSendableTelegramPhotoURL reports whether Telegram sendPhoto is likely to accept the URL.
// Telegram fetches the URL server-side and rejects HTML/SVG and other non-raster types.
func isSendableTelegramPhotoURL(photoURL string) bool {
	photoURL = strings.TrimSpace(photoURL)
	if photoURL == "" {
		return false
	}
	if !strings.HasPrefix(photoURL, "http://") && !strings.HasPrefix(photoURL, "https://") {
		return false
	}
	lower := strings.ToLower(photoURL)
	if strings.Contains(lower, "placehold.co/") {
		// Default placehold.co URLs are SVG; placeholders are not real card art anyway.
		return false
	}
	if strings.HasSuffix(lower, ".svg") {
		return false
	}
	return true
}
