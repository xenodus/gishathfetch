package affiliatelinks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLinkIsActive(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)

	t.Run("active with no expiry", func(t *testing.T) {
		link := Link{Status: StatusActive}
		require.True(t, link.IsActive(now))
	})

	t.Run("inactive status", func(t *testing.T) {
		link := Link{Status: StatusInactive, ExpiryDate: "2099-01-01"}
		require.False(t, link.IsActive(now))
	})

	t.Run("expired link", func(t *testing.T) {
		link := Link{Status: StatusActive, ExpiryDate: "2026-06-28"}
		require.False(t, link.IsActive(now))
	})

	t.Run("expires today", func(t *testing.T) {
		link := Link{Status: StatusActive, ExpiryDate: "2026-06-29"}
		require.True(t, link.IsActive(now))
	})
}
