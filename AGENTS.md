# Agent Instructions

## Priority Order (always apply this tie-breaker)

When trade-offs exist, prioritize in this strict order:

1. Security
2. Data integrity (correct card data)
3. Performance
4. Clean code

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

- For any PR that includes UI changes, include screenshots of the updated UI at both desktop and mobile resolutions.
