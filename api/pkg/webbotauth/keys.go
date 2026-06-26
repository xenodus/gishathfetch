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
// base64-encoded PEM/PKCS8 text.
func ParseEd25519PrivateKeyPEM(raw string) (ed25519.PrivateKey, error) {
	raw = NormalizePrivateKeyMaterial(raw)
	if raw == "" {
		return nil, fmt.Errorf("signing key is empty")
	}

	if key, err := parseEd25519PrivateKeyFromPEMText(raw); err == nil {
		return key, nil
	}

	if !strings.Contains(raw, "BEGIN") {
		der, err := decodeBase64KeyMaterial(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to decode signing key: expected Ed25519 PKCS8 PEM or base64-encoded PKCS8")
		}
		if key, err := parseEd25519PrivateKeyFromDER(der); err == nil {
			return key, nil
		}
		if pemText := string(der); strings.Contains(pemText, "BEGIN") {
			if key, err := parseEd25519PrivateKeyFromPEMText(pemText); err == nil {
				return key, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to decode signing key: expected Ed25519 PKCS8 PEM or base64-encoded PKCS8")
}

func parseEd25519PrivateKeyFromPEMText(pemData string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return parseEd25519PrivateKeyFromDER(block.Bytes)
}

func parseEd25519PrivateKeyFromDER(der []byte) (ed25519.PrivateKey, error) {
	keyIface, err := x509.ParsePKCS8PrivateKey(der)
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
