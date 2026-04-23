---
name: update-libs
description: Update dependencies safely and validate scraper/gateway behavior.
---

Update repository dependencies with risk-aware validation.

Follow this skill document exactly:
- `docs/skills/update-libs.md`

Key must-do constraints:
1. Preserve behavior while upgrading packages using the native package manager.
2. Respect repo priorities for trade-offs:
   - security > correctness > data integrity (correct card data) > performance > clean code.
3. Always run tests before PR updates:
   - `make test`
4. For backend or scraper/gateway touching updates, also run:
   - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
5. Explicitly verify that gateway card search still works for impacted stores, since downstream HTML/API can change unexpectedly.

Return results in this format:
1. `Update Scope`
2. `Changes Made`
3. `Validation`
4. `Risk Notes`
5. `PR`
