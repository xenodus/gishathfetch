package telegramphoto

import (
	"net/url"
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
	if hasUnsupportedImageExtension(lower) {
		return false
	}
	if hasNonStandardPort(photoURL) {
		return false
	}
	return true
}

func hasUnsupportedImageExtension(lowerURL string) bool {
	path := lowerURL
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}
	return strings.HasSuffix(path, ".svg") || strings.HasSuffix(path, ".webp")
}

func hasNonStandardPort(photoURL string) bool {
	parsed, err := url.Parse(photoURL)
	if err != nil {
		return true
	}
	port := parsed.Port()
	if port == "" {
		return false
	}
	return port != "80" && port != "443"
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
