package webbotauth

import (
	"os"
	"path/filepath"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

const testPrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF
-----END PRIVATE KEY-----`

func TestLoadPrivateKeyPEMFromFile(t *testing.T) {
	t.Setenv(config.WebBotAuthPrivateKeyEnv, "")
	t.Setenv(config.WebBotAuthPrivateKeyFileEnv, "")

	keyFile := filepath.Join(t.TempDir(), "web-bot-auth.key")
	require.NoError(t, os.WriteFile(keyFile, []byte(testPrivateKeyPEM), 0o600))
	t.Setenv(config.WebBotAuthPrivateKeyFileEnv, keyFile)

	got, err := LoadPrivateKeyPEM()
	require.NoError(t, err)
	require.Equal(t, testPrivateKeyPEM, got)
}

func TestLoadPrivateKeyPEMFromEnv(t *testing.T) {
	t.Setenv(config.WebBotAuthPrivateKeyFileEnv, "")
	t.Setenv(config.WebBotAuthPrivateKeyEnv, testPrivateKeyPEM)

	got, err := LoadPrivateKeyPEM()
	require.NoError(t, err)
	require.Equal(t, testPrivateKeyPEM, got)
}

func TestPrivateKeyConfigured(t *testing.T) {
	t.Setenv(config.WebBotAuthPrivateKeyEnv, "")
	t.Setenv(config.WebBotAuthPrivateKeyFileEnv, "")
	require.False(t, PrivateKeyConfigured())

	t.Setenv(config.WebBotAuthPrivateKeyEnv, testPrivateKeyPEM)
	require.True(t, PrivateKeyConfigured())
}
