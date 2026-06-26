package webbotauth

import (
	"crypto/ed25519"
	"encoding/base64"
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

func TestParseEd25519PrivateKeyPEM_commonSecretFormats(t *testing.T) {
	pem := `-----BEGIN PRIVATE KEY-----
MC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF
-----END PRIVATE KEY-----`
	singleLine := "-----BEGIN PRIVATE KEY-----MC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF-----END PRIVATE KEY-----"
	escaped := `-----BEGIN PRIVATE KEY-----\nMC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF\n-----END PRIVATE KEY-----`
	quoted := `"-----BEGIN PRIVATE KEY-----\nMC4CAQAwBQYDK2VwBCIEIJ+DYvh6SEqVTm50DFtMDoQikTmiCqirVv9mWG9qfSnF\n-----END PRIVATE KEY-----"`
	base64PEM := base64.StdEncoding.EncodeToString([]byte(pem))

	for _, tc := range []struct {
		name string
		raw  string
	}{
		{name: "multiline pem", raw: pem},
		{name: "single line pem", raw: singleLine},
		{name: "escaped newlines", raw: escaped},
		{name: "quoted escaped pem", raw: quoted},
		{name: "base64 encoded pem", raw: base64PEM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			privateKey, err := ParseEd25519PrivateKeyPEM(tc.raw)
			require.NoError(t, err)
			require.Equal(t, "poqkLGiymh_W0uP6PZFw-dvez3QJT5SolqXBCW38r0U",
				Ed25519JWKThumbprint(privateKey.Public().(ed25519.PublicKey)))
		})
	}
}
