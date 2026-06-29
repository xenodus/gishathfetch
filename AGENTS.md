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
- **Commit** screenshot files to the repo (e.g. under `docs/screenshots/`) and embed them using **`raw.githubusercontent.com` URLs from the PR branch**. Local paths (`/workspace/...`) and Cursor artifact URLs (`cursor.com/artifacts/...`) do not render on GitHub.

  ```markdown
  ![Short description](https://raw.githubusercontent.com/<org>/<repo>/<branch>/docs/screenshots/<file>.png)
  ```

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

- Desktop and mobile full-page screenshots, committed under `docs/screenshots/`
- Include the visible UI changes in each screenshot
- Embed with `raw.githubusercontent.com/<org>/<repo>/<branch>/...` URLs so images render in the PR description

### Known test behaviour

- The `gateway/arcanesanctum` test hits a live website that blocks direct requests without a proxy. It will fail with "Forbidden (proxy_mode=direct proxy=none)" unless `DEDICATED_PROXY_*` env vars are set. This is expected in environments without proxy credentials.
- Many gateway tests hit live upstream store websites, so transient network failures or rate-limiting can cause sporadic failures.

### Frontend API connection

The frontend SPA points to `api.gishathfetch.com` (see `frontend/src/constants.js`). CORS headers on the API allow `gishathfetch.com` and local dev origins (`localhost:5173`). To test the full search flow through the browser, either use the computerUse agent to navigate (it works via fetch from the page) or add a Vite proxy configuration temporarily (revert before committing).

### Backend local mode

When `ENV` is unset (local mode), `go run ./cmd/main.go` executes a hardcoded test search for "Opt" across a subset of stores and prints the JSON result to stdout. No server is started; the process exits after printing.
