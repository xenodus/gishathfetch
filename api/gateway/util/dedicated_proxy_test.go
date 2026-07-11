package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDedicatedProxyReadsEnvVars(t *testing.T) {
	// Set all env vars explicitly to make the test deterministic.
	t.Setenv("DEDICATED_PROXY_1", "host-1|1111|user-1|pass-1")
	t.Setenv("DEDICATED_PROXY_2", "host-2|2222|user-2|pass-2")
	t.Setenv("DEDICATED_PROXY_3", "host-3|3333|user-3|pass-3")
	t.Setenv("DEDICATED_PROXY_4", "host-4|4444|user-4|pass-4")
	t.Setenv("DEDICATED_PROXY_5", "host-5|5555|user-5|pass-5")
	t.Setenv("DEDICATED_PROXY_6", "host-6|6666|user-6|pass-6")
	t.Setenv("DEDICATED_PROXY_7", "host-7|7777|user-7|pass-7")

	got := GetDedicatedProxy()
	require.Len(t, got, 7)

	require.Equal(t, "host-1", got[0].Host)
	require.Equal(t, "1111", got[0].Port)
	require.Equal(t, "user-1", got[0].Username)
	require.Equal(t, "pass-1", got[0].Password)

	require.Equal(t, "host-2", got[1].Host)
	require.Equal(t, "2222", got[1].Port)
	require.Equal(t, "user-2", got[1].Username)
	require.Equal(t, "pass-2", got[1].Password)

	require.Equal(t, "host-3", got[2].Host)
	require.Equal(t, "3333", got[2].Port)
	require.Equal(t, "user-3", got[2].Username)
	require.Equal(t, "pass-3", got[2].Password)

	require.Equal(t, "host-4", got[3].Host)
	require.Equal(t, "4444", got[3].Port)
	require.Equal(t, "user-4", got[3].Username)
	require.Equal(t, "pass-4", got[3].Password)

	require.Equal(t, "host-5", got[4].Host)
	require.Equal(t, "5555", got[4].Port)
	require.Equal(t, "user-5", got[4].Username)
	require.Equal(t, "pass-5", got[4].Password)

	require.Equal(t, "host-6", got[5].Host)
	require.Equal(t, "6666", got[5].Port)
	require.Equal(t, "user-6", got[5].Username)
	require.Equal(t, "pass-6", got[5].Password)

	require.Equal(t, "host-7", got[6].Host)
	require.Equal(t, "7777", got[6].Port)
	require.Equal(t, "user-7", got[6].Username)
	require.Equal(t, "pass-7", got[6].Password)
}

func TestGetDedicatedProxyReturnsEmptyStringsWhenEnvVarsAreEmpty(t *testing.T) {
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}

	got := GetDedicatedProxy()
	require.Len(t, got, 7)
	for _, proxy := range got {
		require.Equal(t, "", proxy.Host)
		require.Equal(t, "", proxy.Port)
		require.Equal(t, "", proxy.Username)
		require.Equal(t, "", proxy.Password)
	}
}

func TestGetDedicatedProxyRejectsPartialSegments(t *testing.T) {
	raw := []string{
		"host-1",
		"host-2|2222",
		"host-3|3333|user-3",
		"host-4|4444|user-4|pass-4|extra",
		"", "", "", "",
	}
	for i, v := range raw {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i+1), v)
	}

	got := GetDedicatedProxy()
	require.Len(t, got, 7)

	for i := range 4 {
		require.Equal(t, DedicatedProxy{}, got[i], "slot %d should reject non-4-segment config", i+1)
	}
}

func TestBuildProxyURL(t *testing.T) {
	t.Run("builds url from pipe separated proxy config", func(t *testing.T) {
		got, ok := BuildProxyURL("dc.com|10000|user-spsykhacft-country-sg|F+password")
		require.True(t, ok)
		require.Equal(t, "http://user-spsykhacft-country-sg:F+password@dc.com:10000", got)
	})

	t.Run("preserves fully formed proxy url", func(t *testing.T) {
		got, ok := BuildProxyURL("http://user:pass@proxy.example:8080")
		require.True(t, ok)
		require.Equal(t, "http://user:pass@proxy.example:8080", got)
	})

	t.Run("rejects incomplete proxy config", func(t *testing.T) {
		got, ok := BuildProxyURL("dc.com")
		require.False(t, ok)
		require.Equal(t, "", got)
	})
}
