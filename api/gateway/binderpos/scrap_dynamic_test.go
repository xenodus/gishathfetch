package binderpos

import (
	"context"
	"errors"
	"testing"

	"mtg-price-checker-sg/gateway"
)

func TestScrapDynamicWithRetryRetriesOn429(t *testing.T) {
	attempts := 0
	cards, err := scrapDynamicWithRetry(context.Background(), func() ([]gateway.Card, error) {
		attempts++
		if attempts < 2 {
			return nil, errors.New("429 Too Many Requests")
		}
		return []gateway.Card{{Name: "ok"}}, nil
	})
	if err != nil {
		t.Fatalf("expected success on second attempt, got %v", err)
	}
	if len(cards) != 1 || cards[0].Name != "ok" {
		t.Fatalf("unexpected cards: %+v", cards)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestScrapDynamicWithRetryDoesNotRetryOnNon429(t *testing.T) {
	attempts := 0
	_, err := scrapDynamicWithRetry(context.Background(), func() ([]gateway.Card, error) {
		attempts++
		return nil, errors.New("403 Forbidden")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestIsHTTPTooManyRequestsRecognizesWrappedScrapError(t *testing.T) {
	err := errors.New("attempt 4 (scrap-dynamic): 429 Too Many Requests (proxy_mode=dynamic proxy=DYNAMIC_PROXY)")
	if !gateway.IsHTTPTooManyRequests(err) {
		t.Fatalf("expected 429 to be detected in %q", err)
	}
}
