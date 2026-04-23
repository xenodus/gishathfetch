# BinderPOS Search Strategy and Retry Summary

Last verified: 2026-04-23

Canonical source files:
- `api/gateway/binderpos/storefront_search.go`
- `api/gateway/binderpos/storefront_fallback.go`
- `api/gateway/binderpos/storefront_client.go`
- `api/gateway/binderpos/storefront_decklist.go`
- `api/gateway/binderpos/storefront_product_details.go`
- `api/gateway/binderpos/scrap.go`
- `api/gateway/collector.go`

## 1) BinderPOS `Search` strategy (current behavior)

Entry point: `binderpos.impl.Search(...)`.

`Search` runs a fixed 3-attempt fallback chain via `searchWithFallback`, and each attempt is wrapped by `runWithAttemptTimeout` with `binderposAttemptTimeout = 2s`.

Attempt order:
1. `api-dedicated`: Storefront API with a dedicated proxy client.
2. `api-shared`: Storefront API with shared proxy (`PROXY_URL`).
3. `scrap-shared`: HTML scraping with shared proxy (`PROXY_URL`), no collector retries.

Per-attempt result handling:
- If an attempt returns `err == nil` (including zero cards), return immediately.
- Fallback to the next strategy only happens when the current attempt returns an error.
- If all attempts fail, return the latest fallback error (priority: attempt 3 -> 2 -> 1).
- Errors are wrapped with attempt labels, e.g. `attempt 2 (api-shared): ...`.

## 2) Storefront API sub-strategy (inside attempts 1 and 2)

`searchByStorefrontAPIWithClient(...)` does:
- Current rollout is 100% decklist endpoint (`binderposDecklistPct = 100`), so `useDecklistForRoll(0..99)` routes to decklist.
- Roll-based selector logic is intentionally preserved for runtime testing/rollback adjustments.

Notes:
- Decklist path depends on host-to-Shopify mapping in `binderposShopifyDomainByStoreHost`; unmapped hosts fail this path and rely on fallback attempts.
- Product-details path code still exists but is currently not selected under the 100% decklist rollout.

## 3) Retry semantics (important distinction)

### A) Retries for `Search`
- `Search` itself has **fallback attempts** (3 sequential strategies), not per-request backoff retries.
- No internal HTTP retry loop is used in attempts 1/2 beyond normal client behavior.
- Attempt 3 explicitly uses a no-retry collector (`NewOptimizedCollectorNoRetry*`), with 2s request timeout.

### B) Retries for direct `Scrap` calls
- `binderpos.impl.Scrap(...)` now uses the no-retry dedicated collector.
- Each scrape request path is single-attempt (no collector retry loop).
- If a scrape attempt returns an error, `Search` fallback logic is responsible for advancing to the next strategy.

## 4) Proxy prerequisites and failure behavior

- Dedicated proxy attempts require at least one configured dedicated proxy URL (`util.GetDedicatedProxyURLs()`); missing/invalid config fails that attempt immediately.
- Shared proxy attempts require `PROXY_URL`; missing/invalid config fails that attempt immediately.
- These config failures are treated as normal fallback errors and drive progression to the next attempt.

## 5) Keep this summary up to date

When changing BinderPOS search/retry logic, update this file in the same PR.

Minimum trigger files:
- `api/gateway/binderpos/storefront_search.go`
- `api/gateway/binderpos/storefront_fallback.go`
- `api/gateway/binderpos/storefront_client.go`
- `api/gateway/binderpos/scrap.go`
- `api/gateway/collector.go` (if BinderPOS retry/proxy flow is affected)
