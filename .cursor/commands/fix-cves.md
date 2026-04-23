---
name: fix-cves
description: Triage and patch open Dependabot CVEs with tested dependency upgrades.
---

Patch open CVEs end-to-end with minimal-risk dependency upgrades.

Use this skill document as the baseline process:
- `docs/skills/fix-cves.md`

Mandatory constraints:
1. Prefer patch/minor upgrades before major upgrades.
2. Never downgrade dependencies to silence advisories.
3. Respect repo trade-off priorities:
   - security > correctness > data integrity (correct card data) > performance > clean code.
4. Run tests before PR handoff:
   - `make test`
5. If backend scraper/gateway dependencies are touched, also run:
   - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
   - and verify gateway card search behavior for impacted stores.

Return results in this format:
1. `Fixed`
2. `Validation`
3. `Risk Notes`
4. `PR`
