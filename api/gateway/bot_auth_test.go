package gateway

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/webbotauth"

	"github.com/stretchr/testify/require"
)

func TestWebBotAuthDisabledByDefault(t *testing.T) {
	t.Setenv(config.WebBotAuthEnabledEnv, "false")
	resetWebBotAuthForTest()

	require.False(t, WebBotAuthEnabled())
	req, err := http.NewRequest(http.MethodGet, "https://shop.example/search", nil)
	require.NoError(t, err)
	require.NoError(t, SignWebBotAuthRequest(req))
	require.Empty(t, req.Header.Get("Signature"))
}

func TestSignWebBotAuthRequestAddsHeaders(t *testing.T) {
	privateKey, pemBytes := generateTestEd25519Key(t)
	t.Setenv(config.WebBotAuthEnabledEnv, "true")
	t.Setenv(config.WebBotAuthPrivateKeyEnv, string(pemBytes))
	t.Setenv(config.WebBotAuthSignatureAgentEnv, "https://gishathfetch.com/.well-known/http-message-signatures-directory")
	resetWebBotAuthForTest()
	t.Cleanup(resetWebBotAuthForTest)

	require.True(t, WebBotAuthEnabled())
	require.Equal(t, defaultBotUserAgent, OutboundUserAgent())

	req, err := http.NewRequest(http.MethodGet, "https://shop.example/search?q=bolt", nil)
	require.NoError(t, err)
	require.NoError(t, SignWebBotAuthRequest(req))

	require.Contains(t, req.Header.Get("Signature-Agent"), "https://gishathfetch.com/.well-known/http-message-signatures-directory")
	require.Contains(t, req.Header.Get("Signature-Input"), `tag="web-bot-auth"`)
	require.Contains(t, req.Header.Get("Signature-Input"), `alg="ed25519"`)
	require.Contains(t, req.Header.Get("Signature"), "sig=:")
	require.NotEmpty(t, req.Header.Get("Signature"))

	thumbprint := webbotauth.Ed25519JWKThumbprint(privateKey.Public().(ed25519.PublicKey))
	require.Contains(t, req.Header.Get("Signature-Input"), thumbprint)
}

func TestPrepareOutboundRequestHTML(t *testing.T) {
	resetWebBotAuthForTest()
	t.Cleanup(resetWebBotAuthForTest)

	req, err := http.NewRequest(http.MethodGet, "https://shop.example/path", nil)
	require.NoError(t, err)
	require.NoError(t, PrepareOutboundRequest(t.Context(), req, OutboundRequestOptions{
		Style:   OutboundStyleHTML,
		PageURL: req.URL,
	}))

	require.NotEmpty(t, req.Header.Get("User-Agent"))
	require.Equal(t, browserLikeAcceptHTML, req.Header.Get("Accept"))
	require.Equal(t, "gzip", req.Header.Get("Accept-Encoding"))
}

func generateTestEd25519Key(t *testing.T) (ed25519.PrivateKey, []byte) {
	t.Helper()
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)
	return privateKey, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func TestMain(m *testing.M) {
	code := m.Run()
	resetWebBotAuthForTest()
	for _, key := range []string{
		config.WebBotAuthEnabledEnv,
		config.WebBotAuthPrivateKeyEnv,
		config.WebBotAuthPrivateKeyFileEnv,
		config.WebBotAuthSignatureAgentEnv,
		config.WebBotAuthUserAgentEnv,
		config.WebBotAuthTTLEnv,
	} {
		_ = os.Unsetenv(key)
	}
	os.Exit(code)
}
