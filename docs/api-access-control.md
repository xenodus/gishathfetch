# API access control (infrastructure)

Production search is **same-origin** on `https://gishathfetch.com`. The browser never calls a public API URL in JavaScript. CloudFront forwards `/api/*` to API Gateway; Lambda enforces an origin secret (phase A) and a short-lived session cookie (phase B).

For scraper timeouts and proxies, see [`search-strategies-retries-timeouts.md`](search-strategies-retries-timeouts.md).

## Public endpoints

| Purpose | URL | Notes |
|---------|-----|--------|
| Search | `GET /api/search?s=...&lgs=...` | Requires valid `gf_api_session` cookie when `API_SESSION_SECRET` is set |
| Session mint | `GET /api/session` | Returns **204** + `Set-Cookie`; called automatically by the SPA |
| SPA | `https://gishathfetch.com/` | Static assets from S3 via CloudFront default behavior |

Query parameters:

- `s` — card name (required, min/max length enforced in Lambda)
- `lgs` — optional comma-separated store names

## Request path (production)

```text
Browser
  → https://gishathfetch.com/api/session   (204, Set-Cookie)
  → https://gishathfetch.com/api/search?... (200, Cookie sent)

CloudFront  (behavior: /api/* → API origin)
  → adds X-Origin-Verify (must match Lambda API_ORIGIN_VERIFY_SECRET)
  → Host: api.gishathfetch.com  (origin domain name)

API Gateway  (HTTP API, e.g. mtg-price-scrapper-API)
  → GET /api/session  → Lambda Session handler
  → GET /api/search   → Lambda Search handler

Lambda  (mtg-price-scrapper)
  → verify X-Origin-Verify
  → verify gf_api_session on search
  → scrape LGS / optional CK lookup
```

Default S3 behavior serves the React app; only paths under `/api/` hit API Gateway.

## AWS components

### API Gateway

Routes on the search HTTP API (auto-deploy to `$default` when enabled):

| Methods | Route | Lambda |
|---------|-------|--------|
| `GET`, `OPTIONS` | `/api/search` | `mtg-price-scrapper` |
| `GET`, `OPTIONS` | `/api/session` | `mtg-price-scrapper` |

Legacy `GET /` on the API custom domain may still exist but is not used by the SPA.

**Custom domain:** `api.gishathfetch.com` maps to this API. It is used as the **CloudFront origin hostname** (CloudFront sends `Host: api.gishathfetch.com`). Do not document it as the public search URL.

**Invoke URL note:** In this project, `https://{api-id}.execute-api.ap-southeast-1.amazonaws.com/...` returns **404** for `/api/*` with the default `Host` header, while the same paths work when `Host` is `api.gishathfetch.com`. CloudFront therefore should **not** use the execute-api hostname as the origin unless that behavior is fixed in API Gateway. CloudFront also **does not allow** a custom origin header named `Host`.

### CloudFront (`gishathfetch.com`)

| Setting | Value |
|---------|--------|
| Behavior path | `/api/*` (above default `*` → S3) |
| Origin domain | `api.gishathfetch.com` |
| Origin path | empty |
| Protocol | HTTPS only, port 443 |
| Cache policy | **CachingDisabled** |
| Origin request policy | **AllViewerExceptHostHeader** (required so API Gateway receives `Host: api.gishathfetch.com`) |
| Compress objects | **No** (recommended for JSON API) |
| Custom origin header | `X-Origin-Verify: <same as API_ORIGIN_VERIFY_SECRET>` |

Phase A blocks direct calls to `api.gishathfetch.com` without the secret (typically **403**). Traffic through CloudFront includes the header.

### Lambda environment (`mtg-price-scrapper`)

| Variable | Purpose |
|----------|---------|
| `API_ORIGIN_VERIFY_SECRET` | Shared secret; must match CloudFront `X-Origin-Verify` |
| `API_ORIGIN_VERIFY_HEADER` | Optional; default `X-Origin-Verify` |
| `API_SESSION_SECRET` | HMAC key for `gf_api_session` cookie |
| `API_SESSION_TTL_SECONDS` | Optional; default **900** (15 minutes) |
| `ENV` | `prod` in production (`Secure` on session cookie) |

If a secret env var is **unset**, that check is disabled (useful for local `go run` and gradual rollout).

**Cookie:** `gf_api_session`, `HttpOnly`, `Path=/api`, `SameSite=Lax`, `Secure` in prod.

## What this mitigates

| Threat | Control |
|--------|---------|
| Other sites embedding the API in browsers | Same-origin `/api/*` + CORS allowlist for dev origins |
| Casual `curl` to `/api/search` without session | Session cookie (phase B) |
| Direct use of API Gateway URL without CloudFront | Origin verify header (phase A) |
| Stale prices at CDN | Caching disabled on `/api/*` |

Does **not** stop a motivated actor replaying requests copied from DevTools (cookie + path). Optional add-ons: **WAF rate limits**, **Turnstile** (not implemented).

## Local development

The Vite dev server proxies `/api` to `VITE_API_PROXY_TARGET` (default `https://api.gishathfetch.com`), preserving paths such as `/api/search` and `/api/session`.

When origin verify is enabled in production, set in `frontend/.env.local` (do not commit):

```env
VITE_API_ORIGIN_VERIFY_SECRET=<same as API_ORIGIN_VERIFY_SECRET>
```

Alternatively proxy to `https://gishathfetch.com` so local traffic matches production edge behavior.

## Verification

```bash
# Session + search via CloudFront (expect 204 then 200)
curl -sS -c /tmp/cj -b /tmp/cj -o /dev/null -w "session: %{http_code}\n" \
  "https://gishathfetch.com/api/session"
curl -sS -b /tmp/cj -o /dev/null -w "search: %{http_code}\n" \
  "https://gishathfetch.com/api/search?s=Opt"

# Direct API host without origin header (expect 403 when phase A enabled)
curl -sS -o /dev/null -w "direct: %{http_code}\n" \
  "https://api.gishathfetch.com/api/search?s=Opt"
```

## Operational pitfalls (learned in rollout)

1. **`AllViewer` origin policy** — forwards `Host: gishathfetch.com` → API **502** or errors; use **AllViewerExceptHostHeader**.
2. **`/api/?s=...` vs `/api/search`** — HTTP API does not accept a route key `/api/` (empty segment). Use **`/api/search`**.
3. **Session 500 after deploy** — same-origin requests often omit `Origin`; ensure Lambda initializes response headers before `Set-Cookie` (see handler `ensureResponseHeaders`).
4. **API Gateway deploy** — routes must be deployed (or auto-deploy on) before `api.gishathfetch.com` paths work.
5. **Removing `api.gishathfetch.com`** — keep it as the CloudFront **origin** hostname until the invoke URL works with CloudFront’s default `Host`, or the site breaks.

## Optional next steps

- **WAF** rate-based rule on CloudFront or API Gateway stage (per-IP abuse).
- **Turnstile** token verification in Lambda before scraping.
- Stop linking `api.gishathfetch.com` publicly; DNS may remain for CloudFront origin resolution.
