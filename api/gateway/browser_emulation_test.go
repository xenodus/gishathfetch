package gateway

import (
	"net/http"
	"testing"
	"time"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestShouldUseBrowserTLSEmulation(t *testing.T) {
	t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "true")
	resetWebBotAuthForTest()
	t.Cleanup(resetWebBotAuthForTest)

	require.True(t, ShouldUseBrowserTLSEmulation(OutboundRequestOptions{}))
	require.True(t, ShouldUseBrowserTLSEmulation(OutboundRequestOptions{SkipWebBotAuth: true}))

	_, pemBytes := generateTestEd25519Key(t)
	t.Setenv(config.WebBotAuthEnabledEnv, "true")
	t.Setenv(config.WebBotAuthPrivateKeyEnv, string(pemBytes))
	t.Setenv(config.WebBotAuthSignatureAgentEnv, "https://gishathfetch.com/.well-known/http-message-signatures-directory")
	resetWebBotAuthForTest()

	require.True(t, WebBotAuthEnabled())
	require.False(t, ShouldUseBrowserTLSEmulation(OutboundRequestOptions{}))
	require.True(t, ShouldUseBrowserTLSEmulation(OutboundRequestOptions{SkipWebBotAuth: true}))

	t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "false")
	require.False(t, ShouldUseBrowserTLSEmulation(OutboundRequestOptions{SkipWebBotAuth: true}))
}

func TestNewOutboundHTTPClient_BrowserTLSByDefault(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "true")
	resetWebBotAuthForTest()
	t.Cleanup(resetWebBotAuthForTest)

	client, err := NewOutboundHTTPClient(2 * time.Second)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.True(t, IsBrowserEmulatedTransport(client.Transport))
}

func TestNewOutboundHTTPClient_StdlibWhenDisabled(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "false")

	client, err := NewOutboundHTTPClient(2 * time.Second)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Nil(t, client.Transport)
}

func TestPickBrowserProfileMatchesUserAgentList(t *testing.T) {
	profile := PickBrowserProfile()
	require.True(t, profile.Enabled)
	require.NotEmpty(t, profile.UserAgent)
	require.Contains(t, browserUserAgents, profile.UserAgent)
}

func TestApplyBrowserProfileHeaders_Chrome(t *testing.T) {
	h := make(map[string][]string)
	headers := (http.Header)(h)
	ApplyBrowserProfileHeaders(&headers, BrowserEmulationProfile{
		Enabled:  true,
		Family:   BrowserFamilyChrome,
		Platform: `"Windows"`,
	})

	require.Equal(t, `"Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`, headers.Get("sec-ch-ua"))
	require.Equal(t, "?0", headers.Get("sec-ch-ua-mobile"))
	require.Equal(t, `"Windows"`, headers.Get("sec-ch-ua-platform"))
	require.Equal(t, "document", headers.Get("sec-fetch-dest"))
}
