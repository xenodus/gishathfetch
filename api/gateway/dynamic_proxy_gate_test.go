package gateway

import (
	"context"
	"testing"
	"time"
)

func TestAcquireDynamicProxySlotLimitsConcurrency(t *testing.T) {
	releases := make([]func(), 0, dynamicProxyMaxConcurrent)
	for i := range dynamicProxyMaxConcurrent {
		release, err := AcquireDynamicProxySlot(context.Background())
		if err != nil {
			t.Fatalf("unexpected error acquiring slot %d: %v", i, err)
		}
		releases = append(releases, release)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := AcquireDynamicProxySlot(ctx); err == nil {
		t.Fatalf("expected acquire to block until ctx deadline when gate is full")
	}

	releases[0]()
	release, err := AcquireDynamicProxySlot(context.Background())
	if err != nil {
		t.Fatalf("expected to acquire a freed slot, got error: %v", err)
	}
	release()

	for _, r := range releases[1:] {
		r()
	}
}

func TestAcquireDynamicProxySlotRespectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := AcquireDynamicProxySlot(ctx); err == nil {
		t.Fatalf("expected cancelled context to return an error")
	}
}
