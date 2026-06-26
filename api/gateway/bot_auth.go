package gateway

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/webbotauth"

	"github.com/gocolly/colly/v2"
	"github.com/lestrrat-go/htmsig"
	"github.com/lestrrat-go/htmsig/component"
	"github.com/lestrrat-go/htmsig/input"
)

const (
	webBotAuthTag                  = "web-bot-auth"
	webBotAuthSignatureAgentMember = "sig"
	defaultBotUserAgent            = "GishathFetch/1.0 (+https://gishathfetch.com; mtg price checker)"
)

type webBotAuthState struct {
	enabled           bool
	privateKey        ed25519.PrivateKey
	keyID             string
	signatureAgentURL string
	userAgent         string
	ttl               time.Duration
}

var (
	loadWebBotAuthOnce sync.Once
	webBotAuth         webBotAuthState
)

// WebBotAuthEnabled reports whether outbound gateway requests should be signed
// per draft-meunier-web-bot-auth-architecture using RFC 9421 HTTP Message Signatures.
func WebBotAuthEnabled() bool {
	loadWebBotAuthOnce.Do(loadWebBotAuth)
	return webBotAuth.enabled
}

// OutboundUserAgent returns a stable bot User-Agent when Web Bot Auth is enabled,
// otherwise a randomized browser User-Agent for legacy scraping behavior.
func OutboundUserAgent() string {
	loadWebBotAuthOnce.Do(loadWebBotAuth)
	if webBotAuth.enabled && webBotAuth.userAgent != "" {
		return webBotAuth.userAgent
	}
	return RandomBrowserUserAgent()
}

// SignWebBotAuthRequest adds Signature, Signature-Input, and Signature-Agent
// headers to req when Web Bot Auth is enabled. It is a no-op when disabled.
func SignWebBotAuthRequest(req *http.Request) error {
	loadWebBotAuthOnce.Do(loadWebBotAuth)
	if !webBotAuth.enabled || req == nil || req.URL == nil {
		return nil
	}

	now := time.Now().Unix()
	nonce, err := newWebBotAuthNonce()
	if err != nil {
		return fmt.Errorf("web bot auth nonce: %w", err)
	}

	req.Header.Set("Signature-Agent", fmt.Sprintf(
		`%s="%s"`,
		webBotAuthSignatureAgentMember,
		webBotAuth.signatureAgentURL,
	))

	sigAgentComponent := component.New("signature-agent").WithParameter("key", webBotAuthSignatureAgentMember)
	def := input.NewDefinitionBuilder().
		Label(webBotAuthSignatureAgentMember).
		Components(component.Authority(), sigAgentComponent).
		Created(now).
		Expires(now + int64(webBotAuth.ttl.Seconds())).
		KeyID(webBotAuth.keyID).
		Algorithm(htmsig.AlgorithmEd25519).
		Nonce(nonce).
		Tag(webBotAuthTag).
		MustBuild()

	ctx := component.WithRequestInfoFromHTTP(context.Background(), req)
	if err := htmsig.SignRequest(ctx, req.Header, input.NewValueBuilder().AddDefinition(def).MustBuild(), webBotAuth.privateKey); err != nil {
		return fmt.Errorf("web bot auth sign request: %w", err)
	}
	return nil
}

// SignWebBotAuthCollyRequest signs a Colly request by synthesizing an http.Request
// for signature generation and copying the resulting headers back.
func SignWebBotAuthCollyRequest(r *colly.Request) error {
	if r == nil || r.URL == nil {
		return nil
	}
	method := strings.TrimSpace(r.Method)
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequest(method, r.URL.String(), nil)
	if err != nil {
		return err
	}
	if r.Headers != nil {
		req.Header = r.Headers.Clone()
	}
	if err := SignWebBotAuthRequest(req); err != nil {
		return err
	}
	if r.Headers == nil {
		h := make(http.Header)
		r.Headers = &h
	}
	for key, values := range req.Header {
		for _, value := range values {
			r.Headers.Set(key, value)
		}
	}
	return nil
}

func loadWebBotAuth() {
	enabled, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(config.WebBotAuthEnabledEnv)))
	if err != nil || !enabled {
		return
	}

	pemData, keyErr := webbotauth.LoadPrivateKeyPEM()
	signatureAgentURL := strings.TrimSpace(os.Getenv(config.WebBotAuthSignatureAgentEnv))
	if keyErr != nil || signatureAgentURL == "" {
		log.Printf(
			"%s is true but signing credentials are not fully configured; Web Bot Auth disabled",
			config.WebBotAuthEnabledEnv,
		)
		return
	}

	privateKey, err := webbotauth.ParseEd25519PrivateKeyPEM(pemData)
	if err != nil {
		log.Printf("invalid Web Bot Auth signing key: %v; Web Bot Auth disabled", err)
		return
	}

	userAgent := defaultBotUserAgent
	if customUA := strings.TrimSpace(os.Getenv(config.WebBotAuthUserAgentEnv)); customUA != "" {
		userAgent = customUA
	}

	keyID := webbotauth.Ed25519JWKThumbprint(privateKey.Public().(ed25519.PublicKey))
	ttl := config.WebBotAuthTTL()
	webBotAuth = webBotAuthState{
		enabled:           true,
		privateKey:        privateKey,
		keyID:             keyID,
		signatureAgentURL: signatureAgentURL,
		userAgent:         userAgent,
		ttl:               ttl,
	}
	log.Printf(
		"Web Bot Auth enabled: signature_agent=%s key_id=%s user_agent=%q ttl=%s",
		signatureAgentURL,
		keyID,
		userAgent,
		ttl,
	)
}

func newWebBotAuthNonce() (string, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// resetWebBotAuthForTest clears cached Web Bot Auth configuration. It is test-only.
func resetWebBotAuthForTest() {
	loadWebBotAuthOnce = sync.Once{}
	webBotAuth = webBotAuthState{}
}
