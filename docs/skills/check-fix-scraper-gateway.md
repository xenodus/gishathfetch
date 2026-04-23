# Skill: check-fix-scraper-gateway

## Purpose

Diagnose and fix scraper/gateway breakages caused by downstream HTML/API changes while preserving result correctness.

## Priority Rule (strict)

When trade-offs appear, enforce:

1. Security
2. Correctness
3. Data integrity (correct card data)
4. Performance
5. Clean code

## Triggers

Use this skill when:

- Store results are empty or unexpectedly low.
- Prices, quantities, set/finish/language fields, or card names are incorrect.
- Search behavior regresses for one or more gateways.
- Scraper tests fail after upstream website/API changes.

## Inputs

- Affected store or gateway package(s) in `api/gateway/...`
- Symptom description (empty result, parse mismatch, timeout, etc.)
- Example search queries/cards for reproduction

## Workflow

1. Reproduce and localize
   - Run focused tests for failing gateway(s).
   - Inspect parser/selectors, normalization logic, and request/response assumptions.

2. Validate data integrity first
   - Confirm card identity fields are correct (name/set/finish/language/collector number when available).
   - Confirm price parsing and currency handling are correct.
   - Confirm deduplication and normalization preserve intended semantics.

3. Patch with minimal scope
   - Update selectors/API mapping and parsing logic only where needed.
   - Avoid broad architecture refactors unless necessary.
   - Preserve clean architecture boundaries (gateway concerns remain in gateway layer).

4. Guard against regressions
   - Add or update tests for the exact breakage.
   - Prefer realistic fixtures/samples that reflect new downstream format.

5. Validate before PR
   - Run repository tests:
     - `make test`
   - Run focused backend suites:
     - `cd api && go test -mod=vendor -failfast -timeout 5m ./gateway/... ./controller/...`
   - Manually sanity-check gateway card search for impacted stores.

## Output checklist

- Clear summary of root cause and impacted store/gateway.
- Data correctness verified for key fields and sorting behavior.
- Tests added/updated and passing.
- Risks or follow-ups documented if downstream format remains unstable.
