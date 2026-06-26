package webbotauth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
)

const directoryMediaType = "application/http-message-signatures-directory+json"

// DirectoryMediaType is the Content-Type for HTTP Message Signatures directories.
func DirectoryMediaType() string {
	return directoryMediaType
}

type directoryKey struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

type directoryDocument struct {
	Keys []directoryKey `json:"keys"`
}

// DirectoryJSON builds the JWKS document for /.well-known/http-message-signatures-directory.
func DirectoryJSON(privateKey ed25519.PrivateKey) ([]byte, error) {
	pub := privateKey.Public().(ed25519.PublicKey)
	doc := directoryDocument{
		Keys: []directoryKey{{
			Kty: "OKP",
			Crv: "Ed25519",
			Kid: Ed25519JWKThumbprint(pub),
			X:   base64.RawURLEncoding.EncodeToString(pub),
			Use: "sig",
			Alg: "ed25519",
		}},
	}
	return json.MarshalIndent(doc, "", "  ")
}
