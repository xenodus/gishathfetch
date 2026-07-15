package gateway

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var httpStatusPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)status[=:\s]+(\d{3})`),
	regexp.MustCompile(`(?i)unexpected status (\d{3})`),
	regexp.MustCompile(`\b(\d{3}) [A-Za-z]`),
}

// FormatHTTPStatus returns a human-readable HTTP status label such as "403 Forbidden".
func FormatHTTPStatus(code int) string {
	if code < 100 || code > 599 {
		return ""
	}

	text := http.StatusText(code)
	if text == "" {
		return fmt.Sprintf("status=%d", code)
	}

	return fmt.Sprintf("%d %s", code, text)
}

// ErrorMessageHasHTTPStatus reports whether msg already includes an HTTP status code.
func ErrorMessageHasHTTPStatus(msg string) bool {
	return ExtractHTTPStatusCode(msg) > 0
}

// IsHTTPServerError reports whether err (or any error in its chain) represents an
// HTTP 5xx response. Bare status phrases such as "Service Unavailable" are
// recognized when they map to a 5xx code.
func IsHTTPServerError(err error) bool {
	for err != nil {
		msg := err.Error()
		code := ExtractHTTPStatusCode(msg)
		if code == 0 {
			code = ExtractHTTPStatusCode(EnsureHTTPStatusInErrorMessage(msg))
		}
		if code >= http.StatusInternalServerError && code < 600 {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

// ExtractHTTPStatusCode returns the first HTTP status code found in msg, or 0.
func ExtractHTTPStatusCode(msg string) int {
	for _, pattern := range httpStatusPatterns {
		matches := pattern.FindStringSubmatch(msg)
		if len(matches) < 2 {
			continue
		}

		code, err := strconv.Atoi(matches[1])
		if err != nil || code < 100 || code > 599 {
			continue
		}

		return code
	}

	return 0
}

// IsHTTPTooManyRequests reports whether err (or any error in its chain) represents
// an HTTP 429 Too Many Requests response.
func IsHTTPTooManyRequests(err error) bool {
	for err != nil {
		code := ExtractHTTPStatusCode(err.Error())
		if code == http.StatusTooManyRequests {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

// EnrichErrorWithHTTPStatus prefixes err with code and status text when statusCode
// is known and the message does not already include a status code.
func EnrichErrorWithHTTPStatus(err error, statusCode int) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	if statusCode < 100 || statusCode > 599 || ErrorMessageHasHTTPStatus(msg) {
		return err
	}

	statusLabel := FormatHTTPStatus(statusCode)
	if statusLabel == "" {
		return err
	}

	if msg == http.StatusText(statusCode) {
		return errors.New(statusLabel)
	}

	return fmt.Errorf("%s: %w", statusLabel, err)
}

// EnsureHTTPStatusInErrorMessage adds a status code to msg when it only contains
// a bare HTTP status phrase such as "Forbidden".
func EnsureHTTPStatusInErrorMessage(msg string) string {
	if ErrorMessageHasHTTPStatus(msg) {
		return msg
	}

	for code := 100; code <= 599; code++ {
		text := http.StatusText(code)
		if text == "" {
			continue
		}

		if msg == text {
			return FormatHTTPStatus(code)
		}

		prefix := text + ": "
		if strings.HasPrefix(msg, prefix) {
			return FormatHTTPStatus(code) + msg[len(text):]
		}

		attemptPrefix := ": " + text
		if idx := strings.LastIndex(msg, attemptPrefix); idx != -1 {
			return msg[:idx+2] + FormatHTTPStatus(code) + msg[idx+2+len(text):]
		}
	}

	return msg
}
