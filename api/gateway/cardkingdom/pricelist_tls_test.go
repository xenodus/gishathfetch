package cardkingdom

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildCKTLSAttempts_IncludesDirectFallback(t *testing.T) {
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	for i := 2; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}
	t.Setenv("DYNAMIC_PROXY", "")

	attempts := buildCKTLSAttempts()
	require.GreaterOrEqual(t, len(attempts), 2)
	require.Equal(t, "dedicated-1", attempts[0].strategy)
	require.Equal(t, "direct", attempts[len(attempts)-1].strategy)
}

func TestLooksLikeCloudflareChallenge(t *testing.T) {
	require.True(t, looksLikeCloudflareChallenge([]byte("<!DOCTYPE html><title>Just a moment...</title>")))
	require.False(t, looksLikeCloudflareChallenge([]byte(`{"meta":{"created_at":"2026-06-28"}}`)))
}
