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

## Required validation before raising a PR

- Run repository tests before opening/updating a PR:
  - `make test`
- For scraper or gateway changes, always run focused backend tests too:
  - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
- Always verify gateway card search behavior for impacted stores. Downstream HTML/API frequently changes due to scraping targets, so regressions may appear even when code compiles.

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

The project requires Go 1.26.3 (per `api/go.mod`). The update script installs it to `/usr/local/go`. You must have `/usr/local/go/bin` in your PATH:

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

### Known test behaviour

- **Live gateway store tests** (`gateway/*/search_test.go`) hit real upstream store websites. They use `gatewaytest.RequireSearchOrProbe`: when search returns cards, field shape is validated; when inventory is empty, tests fall back to HTML/API **structure probes** instead of requiring in-stock results. Transient network failures or rate-limiting can still cause sporadic failures.
- **BinderPOS live integration tests** (`gateway/binderpos/*_test.go`) are skipped by default. Set `RUN_BINDERPOS_LIVE_TESTS=1` to run live storefront/scrape checks against real stores (see also `docs/search-strategies-retries-timeouts.md`).
- **`gateway/arcanesanctum`** is currently skipped because Arcane Sanctum is disabled in the controller; it is not expected to fail on `make test`.
- Some live tests and structure probes may return **403 Forbidden** without `DEDICATED_PROXY_*` credentials when an upstream site blocks direct requests. That is expected in environments without proxy config.

### Frontend API connection

The frontend SPA points to `api.gishathfetch.com` (see `frontend/src/constants.js`). CORS headers on the API allow `gishathfetch.com` and local dev origins (`localhost:5173`). To test the full search flow through the browser, either use the computerUse agent to navigate (it works via fetch from the page) or add a Vite proxy configuration temporarily (revert before committing).

### Backend local mode

When `ENV` is unset (local mode), `go run ./cmd/main.go` executes a hardcoded test search for "Opt" across a subset of stores and prints the JSON result to stdout. No server is started; the process exits after printing.
