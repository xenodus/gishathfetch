package webbotauth

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
)

// ParseEd25519PrivateKeyPEM parses an Ed25519 PKCS8 private key from PEM or
// base64-encoded PEM text.
func ParseEd25519PrivateKeyPEM(raw string) (ed25519.PrivateKey, error) {
	pemData := raw
	if !strings.Contains(raw, "BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, fmt.Errorf("decode base64 private key: %w", err)
		}
		pemData = string(decoded)
	}

	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	privateKey, ok := keyIface.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("expected ed25519 private key, got %T", keyIface)
	}
	return privateKey, nil
}

// Ed25519JWKThumbprint returns the base64url JWK SHA-256 thumbprint for an
// Ed25519 public key, as used for HTTP Message Signatures keyid values.
func Ed25519JWKThumbprint(pub ed25519.PublicKey) string {
	jwk := map[string]string{
		"crv": "Ed25519",
		"kty": "OKP",
		"x":   base64.RawURLEncoding.EncodeToString(pub),
	}
	payload, _ := json.Marshal(jwk)
	sum := sha256.Sum256(payload)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
