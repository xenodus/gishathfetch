package apiauth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"
)

const (
	sessionCookieName = "gf_api_session"
	sessionParts      = 3
)

var (
	// ErrInvalidSessionToken indicates a malformed or tampered session cookie.
	ErrInvalidSessionToken = errors.New("invalid session token")
	// ErrSessionExpired indicates the session cookie is past its expiry.
	ErrSessionExpired = errors.New("session expired")
)

// SessionCookieName is the HttpOnly cookie used for search authorization.
func SessionCookieName() string {
	return sessionCookieName
}

// NewSessionToken mints a signed session value for Set-Cookie.
func NewSessionToken(now time.Time) (string, error) {
	secret := config.APISessionSecret()
	if secret == "" {
		return "", errors.New("session secret not configured")
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	expiry := now.Add(config.APISessionTTL()).Unix()
	payload := fmt.Sprintf("%d.%s", expiry, base64.RawURLEncoding.EncodeToString(nonce))
	sig := signSession(secret, payload)
	return payload + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

// ValidateSessionToken checks the cookie value and returns nil when valid.
func ValidateSessionToken(token string, now time.Time) error {
	secret := config.APISessionSecret()
	if secret == "" {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return ErrInvalidSessionToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != sessionParts {
		return ErrInvalidSessionToken
	}

	expiryUnix, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return ErrInvalidSessionToken
	}
	if now.Unix() > expiryUnix {
		return ErrSessionExpired
	}

	payload := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ErrInvalidSessionToken
	}

	expected := signSession(secret, payload)
	if subtle.ConstantTimeCompare(sig, expected) != 1 {
		return ErrInvalidSessionToken
	}

	return nil
}

func signSession(secret, payload string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}
