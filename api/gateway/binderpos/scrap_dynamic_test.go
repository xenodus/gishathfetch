package binderpos

import (
	"errors"
	"testing"

	"mtg-price-checker-sg/gateway"
)

func TestIsHTTPTooManyRequestsRecognizesWrappedScrapError(t *testing.T) {
	err := errors.New("attempt 4 (scrap-dynamic): 429 Too Many Requests (proxy_mode=dynamic proxy=DYNAMIC_PROXY)")
	if !gateway.IsHTTPTooManyRequests(err) {
		t.Fatalf("expected 429 to be detected in %q", err)
	}
}
