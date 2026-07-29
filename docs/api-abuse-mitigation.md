# API abuse mitigation

The search API is public-facing and relatively expensive (parallel LGS scrapes).
Inbound requests are gated by three optional, layered controls:

1. **CloudFront → API origin secret** (`X-Origin-Verify`)
2. **Browser session cookie** (`gf_api_session`)
3. **Cloudflare Turnstile** on session mint

Each layer is **off until its secret/key is configured**. With none set, `/search`
and `/session` behave as open CORS-allowlisted endpoints (useful for local
Lambda/`go run` testing). In production, enable all three together.

For agent-oriented enablement notes (env vars, Vite, Turnstile checklist), see
also [`AGENTS.md`](../AGENTS.md) → *Frontend API connection*.

---

## Request flow (browser)

```mermaid
sequenceDiagram
    participant U as Browser
    participant TS as Cloudflare Turnstile
    participant API as api.gishathfetch.com
    participant CF_API as Cloudflare siteverify

    Note over U: SPA load (gishathfetch.com)
    U->>TS: invisible widget (when site key set)
    TS-->>U: challenge token
    U->>API: GET /session?cf_turnstile_response=...
    Note over API: origin check (allowlisted Origin<br/>or X-Origin-Verify)
    API->>CF_API: siteverify (when TURNSTILE_SECRET_KEY set)
    CF_API-->>API: success
    API-->>U: 204 + Set-Cookie: gf_api_session=...
    U->>API: GET /search?s=... (credentials: include)
    Note over API: origin check + session cookie HMAC
    API-->>U: search JSON
```

Production SPA calls go **cross-origin** to `https://api.gishathfetch.com`
(`credentials: "include"`). Frontend CDN CloudFront serves static assets from
S3; it is not on the search path unless you also front API Gateway with
CloudFront and inject the origin-verify header (see below).

---

## 1. CloudFront → API origin secret

**Purpose:** Reject callers that are neither an allowlisted browser origin nor a
trusted edge/automation hop that knows the shared secret.

| Item | Value |
|------|--------|
| Lambda env | `API_ORIGIN_VERIFY_SECRET` |
| Optional header name override | `API_ORIGIN_VERIFY_HEADER` (default `X-Origin-Verify`) |
| Code | `api/pkg/apiauth/origin.go`, enforced in `handler/search.go` and `handler/session.go` |

When `API_ORIGIN_VERIFY_SECRET` is set, a request passes origin verification if
**either**:

- The request includes `X-Origin-Verify` (or the configured header) equal to the
  secret — for CloudFront custom origin headers, the Vite dev proxy, or other
  trusted automation; **or**
- The request includes an `Origin` header in the allowlist from
  `config.GetAllowedOrigins()` (`https://gishathfetch.com`,
  `http://localhost:5173`, and the JetBrains built-in server).

Otherwise the handler returns **403** (`forbidden`).

### CloudFront setup (API origin)

If CloudFront is configured as a reverse proxy in front of API Gateway (or as a
custom origin for `/api/*` / search paths), add a **custom origin header**:

- Name: `X-Origin-Verify` (unless you override `API_ORIGIN_VERIFY_HEADER`)
- Value: the same string as Lambda `API_ORIGIN_VERIFY_SECRET`

Keep the secret out of the SPA bundle. Browsers talking directly to
`api.gishathfetch.com` rely on the allowlisted `Origin` path instead of this
header.

### Local Vite proxy

When `VITE_API_BASE_URL=` (empty), Vite proxies `/search` and `/session` to
`VITE_API_PROXY_TARGET`. Set `VITE_API_ORIGIN_VERIFY_SECRET` in
`frontend/.env.local` so the proxy injects `X-Origin-Verify` (see
`frontend/vite.config.js`).

---

## 2. Session token (`gf_api_session`)

**Purpose:** Require a short-lived, HttpOnly cookie before expensive `/search`
work. Anonymous clients must mint a session first; scripts that skip `/session`
cannot search when this layer is on.

| Item | Value |
|------|--------|
| Lambda env | `API_SESSION_SECRET` (HMAC key) |
| Optional TTL | `API_SESSION_TTL_SECONDS` (default **900** = 15 minutes) |
| Cookie name | `gf_api_session` |
| Cookie attrs | `Path=/`, `HttpOnly`, `SameSite=Lax`, `Secure` when `ENV=prod` |
| Code | `api/pkg/apiauth/session.go`, `api/handler/session.go`, `api/handler/access.go` |

### Minting (`GET /session`)

1. Origin verification (layer 1) must pass.
2. `API_SESSION_SECRET` must be set; otherwise **503** (`session not configured`).
3. Turnstile verification runs when configured (layer 3).
4. Response is **204 No Content** with `Set-Cookie: gf_api_session=...` and
   `Cache-Control: no-store`.

Token format: `expiryUnix.nonce.hmac` (HMAC-SHA256 over `expiry.nonce` with
`API_SESSION_SECRET`).

### Enforcement (`GET /search`)

When `API_SESSION_SECRET` is set, `/search` requires a valid cookie:

| Condition | Response body `error` | Status |
|-----------|----------------------|--------|
| Missing / malformed / bad HMAC | `session required` | 403 |
| Past expiry | `session expired` | 403 |

### Frontend behavior

- `ensureApiSession()` (`frontend/src/utils/apiSession.js`) mints the cookie
  before search (and shares one in-flight mint).
- Background refresh every **10 minutes**
  (`API_SESSION_REFRESH_INTERVAL_MS`) so idle tabs stay under the 15-minute TTL.
- On 403 session errors, search retries once after a forced remint.

---

## 3. Cloudflare Turnstile (session mint)

**Purpose:** Slow automated session minting (bots/scripts) before a cookie is
issued. Does not run on `/search` itself — only on `/session`.

| Side | Env / build var | Role |
|------|-----------------|------|
| Lambda | `TURNSTILE_SECRET_KEY` | Cloudflare siteverify; no-op when unset |
| Frontend build | `VITE_TURNSTILE_SITE_KEY` | Loads invisible widget; omit token when unset |

**Enable both together.** Secret-only breaks real browsers; site-key-only adds UI
work with no server check.

### Token transport

The SPA sends the Turnstile token as the **`cf_turnstile_response` query
parameter** on `GET /session` so browsers avoid a CORS preflight. Live API
Gateway `OPTIONS /session` historically did not allow the
`CF-Turnstile-Response` header. Lambda still accepts that header for
compatibility (`api/handler/session_turnstile_token.go`).

### Verification

When `TURNSTILE_SECRET_KEY` is set, Lambda POSTs to Cloudflare
`/turnstile/v0/siteverify` (5s timeout) with the token and optional client IP.
Failures:

| Condition | Response `error` | Status |
|-----------|------------------|--------|
| Token missing | `verification required` | 403 |
| Siteverify reject / HTTP error | `verification failed` | 403 |

### Production enablement checklist

1. Create a Cloudflare Turnstile widget (**Invisible** matches the SPA). Allow
   hostnames `gishathfetch.com` and `localhost` (local Vite).
2. Set `TURNSTILE_SECRET_KEY` on Lambda `mtg-price-scrapper` (not wired by
   `make deploy`).
3. Set `VITE_TURNSTILE_SITE_KEY` at **frontend build** time and redeploy the SPA.
   Production builds may use `frontend/.env.production`; CI must export the key
   if the workflow does not inject it.
4. Confirm `/session` mint succeeds from the live site (cookie set, then
   `/search` works with credentials).

Frontend helpers: `frontend/src/utils/turnstileSession.js`,
`frontend/src/components/TurnstileBootstrap.jsx`.

---

## Environment reference

| Variable | Where | Default / off | Effect when set |
|----------|--------|---------------|-----------------|
| `API_ORIGIN_VERIFY_SECRET` | Lambda | unset = skip | Require `X-Origin-Verify` **or** allowlisted `Origin` |
| `API_ORIGIN_VERIFY_HEADER` | Lambda | `X-Origin-Verify` | Custom header name for the shared secret |
| `API_SESSION_SECRET` | Lambda | unset = skip session on `/search`; `/session` 503 | Sign/validate `gf_api_session` |
| `API_SESSION_TTL_SECONDS` | Lambda | `900` | Cookie / token lifetime |
| `TURNSTILE_SECRET_KEY` | Lambda | unset = skip | Require Turnstile on `/session` |
| `VITE_TURNSTILE_SITE_KEY` | Frontend build | unset = no widget | Invisible Turnstile + token on mint |
| `VITE_API_ORIGIN_VERIFY_SECRET` | Vite only | unset | Proxy injects `X-Origin-Verify` |
| `VITE_API_PROXY_TARGET` | Vite only | `https://api.gishathfetch.com` | Dev proxy target for `/search`, `/session` |
| `VITE_API_BASE_URL` | Frontend | production API host in `constants.js` | Empty string → same-origin paths via Vite proxy |

`config.APIAccessControlEnabled()` is true when origin verify **or** session
secret is configured.

---

## API Gateway routes

Expose **`GET` + `OPTIONS`** for `/search` and `/session` only (legacy `/`,
`/api`, and `/api/*` paths are still accepted by Lambda path routing for
compatibility, but should be removed from the gateway once traffic has
migrated).

CORS allowlist and header policy live in `api/handler/cors.go`.
