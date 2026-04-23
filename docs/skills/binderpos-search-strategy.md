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

`Search` runs a fixed 4-attempt fallback chain via `searchWithFallback`, and each attempt is wrapped by `runWithAttemptTimeout` with `binderposAttemptTimeout = 2s`.

Attempt order:
1. `api-dedicated`: Storefront API with a dedicated proxy client.
2. `api-shared`: Storefront API with shared proxy (`PROXY_URL`).
3. `scrap-dedicated`: HTML scraping with dedicated proxy, no collector retries.
4. `scrap-shared`: HTML scraping with shared proxy (`PROXY_URL`), no collector retries.

Per-attempt result handling:
- If an attempt returns `err == nil` and `len(cards) > 0`, return immediately.
- If an attempt returns `err == nil` and zero cards, continue to next attempt.
- If at least one attempt ended with `err == nil` (even with zero cards), final result is `[]` with `nil` error.
- If all attempts fail, return the latest fallback error (priority: attempt 4 -> 3 -> 2 -> 1).
- Errors are wrapped with attempt labels, e.g. `attempt 2 (api-shared): ...`.

## 2) Storefront API sub-strategy (inside attempts 1 and 2)

`searchByStorefrontAPIWithClient(...)` does:
- 70% path: BinderPOS decklist endpoint (`useDecklistForRoll(0..69)`).
- 30% path: Shopify suggest + per-product details fallback (`useDecklistForRoll(70..99)`).

Notes:
- Decklist path depends on host-to-Shopify mapping in `binderposShopifyDomainByStoreHost`; unmapped hosts fail this path and rely on fallback attempts.
- Product-details path tolerates individual product-detail fetch failures (skips failed products) and can still return partial success with `nil` error.

## 3) Retry semantics (important distinction)

### A) Retries for `Search`
- `Search` itself has **fallback attempts** (4 sequential strategies), not per-request backoff retries.
- No internal HTTP retry loop is used in attempts 1/2 beyond normal client behavior.
- Attempts 3/4 explicitly use no-retry collectors (`NewOptimizedCollectorNoRetry*`), with 2s request timeout.

### B) Retries for direct `Scrap` calls
- `binderpos.impl.Scrap(...)` (not `Search`) uses `NewOptimizedCollectorForBinderpos(...)`, which enables collector retries.
- Current collector retry policy: `maxRetries = 2` (two retries after initial request).
- Proxy progression for direct `Scrap` retries:
  - Default: `dedicated -> dedicated -> direct`.
  - Rollback mode (`USE_BINDERPOS_SHARED_PROXY_FALLBACK=true`): `dedicated -> shared(PROXY_URL) -> direct`.

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
