# Agent Instructions

## Priority Order (always apply this tie-breaker)

When trade-offs exist, prioritize in this strict order:

1. Security
2. Correctness
3. Data integrity (correct card data)
4. Performance
5. Clean code

Notes:
- Correctness = expected functional behavior (search logic, filtering, sorting, and API behavior).
- Data integrity = accuracy and consistency of card fields (name/set/finish/language/price).

## Backend (Go) implementation standards

- Backend work in `api/` must follow clean code and clean architecture boundaries.
- Keep domain logic and gateway/scraper concerns separated; avoid shortcut coupling across layers.
- Favor small, testable changes over broad refactors unless the task explicitly requires larger restructuring.

## End of session

- Before finishing every session, run `go fix` on the backend:
  - `cd api && go fix ./...`
- If `go fix` changes any files, include those changes in the same commit/PR as the session work.

## Required validation before raising a PR

- Run repository tests before opening/updating a PR:
  - `make test`
- For scraper or gateway changes, always run focused backend tests too:
  - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
- Always verify gateway card search behavior for impacted stores. Downstream HTML/API frequently changes due to scraping targets, so regressions may appear even when code compiles.

## Privacy policy

After each change (before opening or updating a PR), check whether the privacy policy
needs updating. Do this even when the task is not privacy-related.

### Where policies live

| Policy | Source of truth | Public URL |
|--------|-----------------|------------|
| Main website | `frontend/src/components/Modals.jsx` (Privacy modal) | [https://gishathfetch.com/?privacy=1](https://gishathfetch.com/?privacy=1) |
| Telegram bot | `frontend/public/telegram-bot-privacy.html` | [https://gishathfetch.com/telegram-bot-privacy.html](https://gishathfetch.com/telegram-bot-privacy.html) |

Edit the relevant source file directly when privacy text changes. Cross-link between the
main policy and the Telegram policy when either is edited.

### When to update

Review privacy impact when a change introduces or alters any of:

- Cookies or browser storage (for example `gf_api_session`, theme, saved cards, favourites)
- Analytics or telemetry (GA4 events, search terms, Trending, server-side measurement)
- Advertising (AdSense or similar third-party scripts)
- New user-facing channels (Telegram, email, other bots or integrations)
- Personal or identifiable data collection, logging, or retention
- Third-party services that receive user input or identifiers (for example Scryfall autocomplete)

If disclosure is needed, update the relevant policy in the same PR. Note what changed
in the PR description. Refresh privacy-modal screenshots when visible policy text changes.

If no update is required, say so briefly in the PR (for example: "Privacy policy:
no change — no new data collection").

## UI deliverables

For any PR that includes UI changes:

- Include screenshots in the PR description at both **desktop** and **mobile** resolutions.
- **Take full-page screenshots** that include the visible UI changes. Avoid tight crops of a single component; reviewers should see the change in the context of the full page.
- **Refresh screenshots whenever further visible UI changes are made** on the same PR. Do not leave stale screenshots in the PR description after follow-up styling or layout edits.
- **Avoid cached screenshots** when updating a PR:
  - Save each new capture under a **new filename** (for example, append a date or version suffix like `homepage-search-20260629-v2-desktop.png`).
  - Update the PR description to reference the new image paths so GitHub uploads fresh assets instead of reusing old URLs.
  - Do not overwrite an existing screenshot filename if the PR already references it.
- **Do not commit screenshot files to the repo.** Screenshots are PR-only artifacts used in the PR description.
- Embed screenshots using **GitHub-hosted image URLs that render in PR descriptions** (for example `github.com/.../releases/download/...` or `github.com/user-attachments/assets/...`). Do not use `cursor.com/artifacts/...` URLs or uncommitted local file paths.

## Cursor Cloud specific instructions

### Services overview

| Service | Location | Run command | Port |
|---------|----------|-------------|------|
| Go backend (Lambda handler) | `api/` | `cd api && go run -mod=vendor ./cmd/main.go` | N/A (one-shot, prints JSON) |
| Frontend dev server (Vite) | `frontend/` | `cd frontend && npm run dev` | 5173 |

### Go version requirement

The project requires Go 1.27.0 (per `api/go.mod`). Cloud Agent VMs install the matching toolchain to `/usr/local/go` via `.cursor/environment.json` → `.cursor/scripts/cloud-agent-install.sh` (reads `api/go.mod`). You must have `/usr/local/go/bin` in your PATH:

```bash
export PATH="/usr/local/go/bin:$PATH"
```

### Running tests

- Full test suite: `make test` (from repo root)
- Gateway/controller focused: `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
- Frontend lint: `cd frontend && npm run lint`

### UI screenshots

Follow the [UI deliverables](#ui-deliverables) rules above. In short:

- Desktop and mobile full-page screenshots with the visible UI changes
- Re-capture and update PR screenshots after any follow-up visible UI change on the same PR
- Use a new filename for each refresh so PR images are not served from cache
- Do not commit screenshots to the repo; attach them only in the PR description
- Use GitHub-hosted image URLs that render on GitHub (not Cursor artifact URLs)

#### Generating screenshots (pre-installed tooling)

A reusable full-page screenshot helper is pre-provisioned at `~/.agent-tools/screenshots/screenshot.mjs`. It drives the system-installed Google Chrome via Playwright (no bundled-browser download; the startup update script keeps its deps installed).

Start the dev server first (`cd frontend && npm run dev`, port 5173), then run:

```bash
node ~/.agent-tools/screenshots/screenshot.mjs http://localhost:5173 /opt/cursor/artifacts homepage
```

This writes `homepage-desktop.png` (1440x900, full page) and `homepage-mobile.png` (iPhone 13 emulation, full page). Pass `--desktop-only` or `--mobile-only` as a 4th arg to capture just one. Point the URL at any route/query (e.g. `'http://localhost:5173/?s=Opt'`) to capture a specific state. Screenshots are dev-only artifacts — do not commit them (see the caching/GitHub-URL rules above).

### Known test behaviour

- **Live gateway store tests** (`gateway/*/search_test.go`) hit real upstream store websites. They use `gatewaytest.RequireSearchOrProbe`: when search returns cards, field shape is validated; when inventory is empty, tests fall back to HTML/API **structure probes** instead of requiring in-stock results. Transient network failures or rate-limiting can still cause sporadic failures.
- **BinderPOS live integration tests** (`gateway/binderpos/*_test.go`) are skipped by default. Set `RUN_BINDERPOS_LIVE_TESTS=1` to run live storefront/scrape checks against real stores (see also `docs/search-strategies-retries-timeouts.md`).
- Some live tests and structure probes may return **403 Forbidden** without `DEDICATED_PROXY_*` credentials when an upstream site blocks direct requests. That is expected in environments without proxy config.

### Frontend API connection

Production search uses **`https://api.gishathfetch.com/search`** and **`/session`** (see
`frontend/src/constants.js`). Set `VITE_API_BASE_URL=` (empty) in `frontend/.env.local` to use
Vite proxies on `/search` and `/session` during local dev; otherwise the SPA calls the API host
directly (CORS allowlist includes `http://localhost:5173`).

Proxy target: `VITE_API_PROXY_TARGET` (default `https://api.gishathfetch.com`). Set
`VITE_API_ORIGIN_VERIFY_SECRET` when testing against an environment with
`API_ORIGIN_VERIFY_SECRET` enabled.

Full reference for the two abuse-mitigation layers:
[`docs/api-abuse-mitigation.md`](docs/api-abuse-mitigation.md).

#### API abuse mitigation (overview)

Two optional layers; each is **off until configured**:

1. **CloudFront → API origin secret** (`API_ORIGIN_VERIFY_SECRET`): when set, requests
   must include a matching `X-Origin-Verify` header injected by the API CloudFront
   distribution (origin request) or the Vite dev proxy. Allowlisted `Origin` alone
   is not accepted. Enforced on both `/session` and `/search`.
2. **Session cookie** (`API_SESSION_SECRET`): `GET /session` mints HttpOnly `gf_api_session`
   (default TTL 15m via `API_SESSION_TTL_SECONDS`). `/search` requires a valid cookie
   (`session required` / `session expired` → 403). Frontend: `ensureApiSession()` with a
   10-minute background refresh and one retry on expiry
   (`frontend/src/utils/apiSession.js`).

Production should enable both together. Both public CloudFront distributions
(`gishathfetch.com` and `api.gishathfetch.com`) have **AWS WAF** web ACLs
attached; see [`docs/api-abuse-mitigation.md`](docs/api-abuse-mitigation.md) →
*Edge protection*. `make deploy` does not wire these Lambda
secrets — set them on `mtg-price-scrapper` manually (or via your secrets process).

**API Gateway (manual):** expose `GET` + `OPTIONS` for `/search` and `/session` only; remove
legacy `/`, `/api`, and `/api/*` routes when traffic has migrated.

### Backend local mode

When `ENV` is unset (local mode), `go run ./cmd/main.go` executes a hardcoded test search for "Opt" across a subset of stores and prints the JSON result to stdout. No server is started; the process exits after printing.
