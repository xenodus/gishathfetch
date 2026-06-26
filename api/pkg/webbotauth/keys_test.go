package webbotauth

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEd25519JWKThumbprintMatchesRFC9421TestKey(t *testing.T) {
	privateKey, err := ParseEd25519PrivateKeyPEM(`-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF
-----END PRIVATE KEY-----`)
	require.NoError(t, err)

	thumbprint := Ed25519JWKThumbprint(privateKey.Public().(ed25519.PublicKey))
	require.Equal(t, "poqkLGiymh_W0uP6PZFw-dvez3QJT5SolqXBCW38r0U", thumbprint)
}

func TestDirectoryJSON(t *testing.T) {
	privateKey, err := ParseEd25519PrivateKeyPEM(`-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF
-----END PRIVATE KEY-----`)
	require.NoError(t, err)

	body, err := DirectoryJSON(privateKey)
	require.NoError(t, err)
	require.Contains(t, string(body), `"kid": "poqkLGiymh_W0uP6PZFw-dvez3QJT5SolqXBCW38r0U"`)
	require.Contains(t, string(body), `"alg": "ed25519"`)
	require.Contains(t, string(body), `"use": "sig"`)
}
