package apiauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"
)

const (
	// TurnstileResponseHeader is the request header carrying the browser Turnstile token.
	TurnstileResponseHeader = "CF-Turnstile-Response"
	// TurnstileResponseQueryParam is an alternate token carrier for GET /session so browsers
	// avoid a CORS preflight when API Gateway does not echo Allow-Headers for the header above.
	TurnstileResponseQueryParam = "cf_turnstile_response"
	turnstileSiteverifyURLDefault = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	turnstileVerifyTimeout  = 5 * time.Second
)

var (
	// ErrTurnstileTokenMissing indicates the client did not send a Turnstile token.
	ErrTurnstileTokenMissing = errors.New("turnstile token missing")
	// ErrTurnstileVerificationFailed indicates Cloudflare rejected the token.
	ErrTurnstileVerificationFailed = errors.New("turnstile verification failed")
)

type turnstileSiteverifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

var (
	turnstileHTTPClient   = &http.Client{Timeout: turnstileVerifyTimeout}
	turnstileSiteverifyURL = turnstileSiteverifyURLDefault
)

// VerifyTurnstile checks the token with Cloudflare when TURNSTILE_SECRET_KEY is set.
//
// siteverify's optional remoteip field is intentionally not sent: API Gateway
// SourceIP can disagree with the address Cloudflare observed for the challenge
// (IPv4/IPv6 dual-stack and privacy/VPN egress are common in incognito), which
// caused intermittent invalid-input-response failures. Token cryptographic
// validation is sufficient for session mint abuse mitigation.
func VerifyTurnstile(ctx context.Context, token string) error {
	secret := config.TurnstileSecretKey()
	if secret == "" {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTurnstileTokenMissing
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		turnstileSiteverifyURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := turnstileHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("turnstile siteverify: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("turnstile response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return ErrTurnstileVerificationFailed
	}

	var parsed turnstileSiteverifyResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("turnstile decode: %w", err)
	}
	if !parsed.Success {
		return ErrTurnstileVerificationFailed
	}

	return nil
}
