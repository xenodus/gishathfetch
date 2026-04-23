# Skill: Update Libraries

## Objective

Upgrade project dependencies safely while preserving behavior and keeping scraper reliability high.

## Priority Rules

Apply this order when trade-offs exist:

1. Data integrity (correct card data)
2. Performance
3. Clean code

## Scope

- Go backend deps (`api/go.mod`, `api/go.sum`, `api/vendor/`)
- Frontend deps (`frontend/package.json`, lockfile if present)
- Build/runtime references (Dockerfile, CI/workflows, scripts)

## Workflow

1. Discover outdated dependencies:
   - Go: `cd api && go list -m -u all`
   - Frontend: `cd frontend && npm outdated` (if frontend scope is included)
2. Build an upgrade plan:
   - Prefer patch/minor upgrades first.
   - Group upgrades by ecosystem and risk.
   - Mark any major upgrades with expected behavior impact.
3. Implement:
   - Use native package managers (`go get`, `npm install`/`npm update`).
   - Regenerate dependency artifacts with official tooling:
     - Go: `go mod tidy && go mod vendor`
4. Validate:
   - Always run full tests: `make test`
   - For any gateway/scraper-related dependency changes:
     - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
   - Manually verify gateway card search behavior for affected stores.
5. Deliver:
   - Summarize changed packages and old -> new versions.
   - Call out risky upgrades and follow-up checks.

## Done Criteria

- Dependency files are consistent and committed.
- Tests pass with required scraper/gateway validations.
- Any residual upgrade risks are documented.
