package agora

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgoraOutboundOpts_ProductionHost(t *testing.T) {
	pageURL, err := url.Parse(StoreBaseURL + StoreSearchPath)
	require.NoError(t, err)

	opts := agoraOutboundOpts(pageURL)
	require.True(t, opts.SkipWebBotAuth)
	require.True(t, opts.SkipDirect)
	require.True(t, opts.PreferResidentialProxy)
}

func TestAgoraOutboundOpts_NonProductionHostAllowsDirect(t *testing.T) {
	pageURL, err := url.Parse("http://127.0.0.1:0/store/search")
	require.NoError(t, err)

	opts := agoraOutboundOpts(pageURL)
	require.True(t, opts.SkipWebBotAuth)
	require.False(t, opts.SkipDirect)
	require.True(t, opts.PreferResidentialProxy)
}
