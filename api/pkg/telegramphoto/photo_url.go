package telegramphoto

import (
	"net/url"
	"strings"
)

// Telegram sendPhoto via URL accepts JPEG, PNG, and GIF raster images.
var allowedImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif"}

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
// Telegram fetches the URL server-side and only reliably accepts JPEG, PNG, and GIF.
func IsSendable(photoURL string) bool {
	photoURL = strings.TrimSpace(photoURL)
	if photoURL == "" {
		return false
	}
	if !strings.HasPrefix(photoURL, "http://") && !strings.HasPrefix(photoURL, "https://") {
		return false
	}
	if !hasAllowedImageExtension(photoURL) {
		return false
	}
	if hasNonStandardPort(photoURL) {
		return false
	}
	return true
}

func hasAllowedImageExtension(photoURL string) bool {
	path := strings.ToLower(photoURL)
	if i := strings.IndexAny(path, "?#"); i >= 0 {
		path = path[:i]
	}
	for _, ext := range allowedImageExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
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
