package adminauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SessionCookieName = "gishathfetch_admin_session"

var (
	ErrInvalidSession = errors.New("invalid session")
	ErrExpiredSession = errors.New("expired session")
)

type sessionPayload struct {
	Username  string `json:"u"`
	ExpiresAt int64  `json:"e"`
}

func IssueSessionToken(secret, username string, ttl time.Duration) (string, error) {
	secret = strings.TrimSpace(secret)
	username = strings.TrimSpace(username)
	if secret == "" || username == "" {
		return "", fmt.Errorf("adminauth: missing session inputs")
	}

	payload := sessionPayload{
		Username:  username,
		ExpiresAt: time.Now().UTC().Add(ttl).Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	payloadPart := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := sign(secret, payloadPart)
	return payloadPart + "." + signature, nil
}

func ValidateSessionToken(secret, token string) (string, error) {
	secret = strings.TrimSpace(secret)
	token = strings.TrimSpace(token)
	if secret == "" || token == "" {
		return "", ErrInvalidSession
	}

	payloadPart, signature, ok := strings.Cut(token, ".")
	if !ok || payloadPart == "" || signature == "" {
		return "", ErrInvalidSession
	}

	expectedSignature := sign(secret, payloadPart)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
		return "", ErrInvalidSession
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return "", ErrInvalidSession
	}

	var payload sessionPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return "", ErrInvalidSession
	}
	if strings.TrimSpace(payload.Username) == "" {
		return "", ErrInvalidSession
	}
	if time.Now().UTC().Unix() > payload.ExpiresAt {
		return "", ErrExpiredSession
	}

	return payload.Username, nil
}

func CredentialsMatch(expectedUsername, expectedPassword, providedUsername, providedPassword string) bool {
	usernameMatches := subtle.ConstantTimeCompare(
		[]byte(strings.TrimSpace(expectedUsername)),
		[]byte(strings.TrimSpace(providedUsername)),
	) == 1
	passwordMatches := subtle.ConstantTimeCompare(
		[]byte(strings.TrimSpace(expectedPassword)),
		[]byte(strings.TrimSpace(providedPassword)),
	) == 1
	return usernameMatches && passwordMatches
}

func SessionCookie(token string, maxAgeSeconds int) string {
	return fmt.Sprintf(
		"%s=%s; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=%d",
		SessionCookieName,
		token,
		maxAgeSeconds,
	)
}

func ClearSessionCookie() string {
	return fmt.Sprintf("%s=; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=0", SessionCookieName)
}

func SessionTokenFromCookie(cookieHeader string) string {
	for part := range strings.SplitSeq(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		name, value, found := strings.Cut(part, "=")
		if found && name == SessionCookieName {
			return value
		}
	}
	return ""
}

func sign(secret, payloadPart string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payloadPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
