package apiauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"
)

var turnstileSiteVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

// ErrTurnstileVerificationFailed indicates the Turnstile token was missing or rejected.
var ErrTurnstileVerificationFailed = errors.New("turnstile verification failed")

type turnstileSiteVerifyResponse struct {
	Success  bool   `json:"success"`
	Hostname string `json:"hostname"`
}

var turnstileHTTPClient = &http.Client{Timeout: 5 * time.Second}

// allowedTurnstileHostnames lists page hostnames where the Turnstile widget may run.
// Tokens are minted on the SPA origin (gishathfetch.com), not api.gishathfetch.com.
func allowedTurnstileHostnames() []string {
	hostnames := []string{"gishathfetch.com"}
	if os.Getenv("ENV") != config.EnvProd {
		hostnames = append(hostnames, "localhost")
	}
	return hostnames
}

func isAllowedTurnstileHostname(hostname string) bool {
	hostname = strings.ToLower(strings.TrimSpace(hostname))
	if hostname == "" {
		return false
	}
	return slices.Contains(allowedTurnstileHostnames(), hostname)
}

// VerifyTurnstileToken checks a browser Turnstile response with Cloudflare when configured.
// When TURNSTILE_SECRET_KEY is unset, verification is skipped.
func VerifyTurnstileToken(ctx context.Context, token, remoteIP string) error {
	secret := config.TurnstileSecretKey()
	if secret == "" {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTurnstileVerificationFailed
	}

	form := url.Values{}
	form.Set("secret", secret)
	form.Set("response", token)
	if remoteIP = strings.TrimSpace(remoteIP); remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		turnstileSiteVerifyURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := turnstileHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}

	var parsed turnstileSiteVerifyResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return err
	}
	if !parsed.Success {
		return ErrTurnstileVerificationFailed
	}
	if !isAllowedTurnstileHostname(parsed.Hostname) {
		return ErrTurnstileVerificationFailed
	}

	return nil
}
