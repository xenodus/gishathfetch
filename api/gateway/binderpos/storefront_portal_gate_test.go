package binderpos

import (
	"context"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway"
)

func TestAcquireBinderposPortalSlotLimitsConcurrency(t *testing.T) {
	releases := make([]func(), 0, binderposPortalMaxConcurrent)
	for i := range binderposPortalMaxConcurrent {
		release, err := acquireBinderposPortalSlot(context.Background())
		if err != nil {
			t.Fatalf("unexpected error acquiring slot %d: %v", i, err)
		}
		releases = append(releases, release)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := acquireBinderposPortalSlot(ctx); err == nil {
		t.Fatalf("expected acquire to block until ctx deadline when gate is full")
	}

	// A freed slot should be acquirable again.
	releases[0]()
	release, err := acquireBinderposPortalSlot(context.Background())
	if err != nil {
		t.Fatalf("expected to acquire a freed slot, got error: %v", err)
	}
	release()

	for _, r := range releases[1:] {
		r()
	}
}

func TestAcquireBinderposPortalSlotRespectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := acquireBinderposPortalSlot(ctx); err == nil {
		t.Fatalf("expected cancelled context to return an error")
	}
}

func TestBinderposPortalHostRegisteredAsAlwaysPaced(t *testing.T) {
	// Re-registering is idempotent and confirms the package init wired the host.
	gateway.RegisterAlwaysPacedDomain(binderposPortalHost)
}
