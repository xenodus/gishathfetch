---
name: check-fix-scraper-gateway
description: Diagnose and fix scraper/gateway card search regressions caused by downstream site changes.
---

Diagnose and fix scraper/gateway card search regressions end-to-end, with card-data correctness as the top priority.

Use this skill document as the operational checklist:

- `docs/skills/check-fix-scraper-gateway.md`

Non-negotiable priorities:

1. Data integrity (correct card data)
2. Performance
3. Clean code / architecture

Before PR handoff, you must:

1. Run focused gateway/controller tests:
   - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
2. Run repository test baseline:
   - `make test`
3. Report what stores/queries were validated and any known scrape fragility that remains.
