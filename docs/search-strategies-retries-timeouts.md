# Search strategies, retries, and timing

This document records **where** the app configures search behavior, **timeouts**, **fallback/attempt ordering**, **concurrency limits**, and **jittered pacing**. It is meant for code agents and maintainers: when you change a constant, update this file in the same PR.

---

## Backend: controller (multi-store search)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Per-store deadline | 20s | `config.PerSiteTimeout` in `api/pkg/config/config.go`; used in `searchShop` as `context.WithTimeout` in `api/controller/search.go` | One goroutine per selected store; each `LGS.Search` runs under this cap. |
| Colly request timeout (default scrapers) | 20s | `applyCollectorDefaults` → `c.SetRequestTimeout(config.PerSiteTimeout)` in `api/gateway/collector.go` | Overrides gocolly’s default 10s for optimized collectors. |
| Minimum end-to-end response time | 1s | `responseThreshold` in `searchShops` in `api/controller/search.go` | If all stores finish in under 1s, the handler **sleeps** the remainder so the API “feels” less instant. |
| Colly HTTP retries | None | `api/gateway/collector.go` (`configureRequestOptimizations`, `registerNoRetryErrorHandler`) | **Single HTTP attempt** per colly request path; no automatic colly/gateway retry of failed visits. |
| BinderPOS store concurrency | 12 | `binderposMaxConcurrent` in `api/controller/search.go` | Semaphore limits how many binderpos-backed shops run at once. Non-binderpos stores are not limited by this gate. |

---

## Backend: domain request pacing (all colly + `WaitForDomainRequestSlot` users)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Minimum interval between requests to the **same host** | 300ms | `domainRequestMinInterval` in `api/gateway/domain_rate_limiter.go` | Added to the reservation for the next allowed time for that domain. |
| Jitter (added on top of minimum interval) | uniform in **[0, 600ms)** | `domainRequestMaxJitter` + `randomDuration` in `api/gateway/domain_rate_limiter.go` | Per reservation: `reservedUntil = nextAllowed + minInterval + jitter`. If the wait is cancelled, the limiter can roll back that reservation. |

## Backend: dedicated proxy env (`api/gateway/util/dedicated_proxy.go`)

| Item | Value | Notes |
|------|--------|--------|
| Configured slots | **`DEDICATED_PROXY_1`** … **`DEDICATED_PROXY_7`** | Each value is `host\|port\|username\|password` (pipe-separated). Empty or incomplete entries are ignored when building URLs. |
| Dynamic fallback proxy | **`DYNAMIC_PROXY`** | Uses the same `host\|port\|username\|password` format as `DEDICATED_PROXY_*` (full proxy URLs are also accepted). BinderPOS reserves it for the final fallback after dedicated and direct/no-proxy attempts. |

---

## Live BinderPOS integration tests

Some tests in `api/gateway/binderpos/*_test.go` hit real stores and proxies. They run only when **`RUN_BINDERPOS_LIVE_TESTS=1`** is set (default `make test` skips them to avoid rate limits and flaky remote dependencies).

---

## Backend: BinderPOS (storefront + scraper fallbacks)

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| `shopifyDomain` **non-empty** (normal) | 1) **api-dedicated** → 2) **api-direct** → 3) **scrap-dedicated** → 4) **scrap-direct** → 5) **api-dynamic** → 6) **scrap-dynamic** | **10s** per step: `binderposAttemptTimeout` in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. HTTP clients in `storefront_client.go` use the same `binderposAttemptTimeout`. |
| `shopifyDomain` **empty** | **scrap-dedicated** → **scrap-direct** → **scrap-dynamic** (three `runWithAttemptTimeout` steps in `searchWithScrapDedicatedThenDirectThenDynamic`) | **10s** per step (same constant). No API attempts. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| **api-dedicated** proxy selection (storefront HTTP client) | Round-robin | `nextBinderposStorefrontProxyURL` + `binderposDedicatedProxySeq` in `api/gateway/binderpos/storefront_client.go` | When `UseLeasedDedicatedProxy` is **false** (default in `api/pkg/config/config.go`), each storefront API call cycles **proxy₁ → … → proxyₙ** then repeats, using all URLs from `GetDedicatedProxyURLs()`. When `UseLeasedDedicatedProxy` is **true**, selection is unchanged: `LeaseDedicatedProxyURL` from the dedicated pool. |
| **api-dynamic** proxy selection (storefront HTTP client) | Fixed env URL | `searchByStorefrontAPIDynamic` in `api/gateway/binderpos/storefront_client.go` | Uses `DYNAMIC_PROXY` as the authenticated proxy URL for the final BinderPOS API fallback attempt. |
| Colly for BinderPOS scrapes | 10s | `SetRequestTimeout(binderposAttemptTimeout)` in `api/gateway/binderpos/scrap.go` | Tighter than generic `PerSiteTimeout` (20s) for binderpos scrape collectors. |
| “Retries” | N/A (sequential fallbacks) | `searchWithFallback` | Stops on first **success**; returns last error if all attempts fail. This is **not** exponential backoff retry of a single request. |

---

## Frontend: `useSearch` (API + Scryfall)

Constants live in `frontend/src/hooks/useSearch.js` (and related).

| Item | Value | Notes |
|------|--------|--------|
| Autocomplete debounce | 300ms | `AUTOCOMPLETE_DEBOUNCE_MS`; delays Scryfall autocomplete fetches after typing. |
| Search progress UI tick | 1000ms | `SEARCH_PROGRESS_INTERVAL_MS`; animates the “Searching LGS . . .” label. |
| Programmatic search delay on load (URL with `?s=`) | 100ms | `setTimeout` before `performSearch` in the mount `useEffect`. |
| API `fetch` timeout / retries | None in code | Uses browser `fetch` with `AbortController` only; no app-level timeout or automatic retry. |

---

## How to keep this file accurate

1. When adding or changing **timeouts, intervals, concurrency, or strategy order**, update the relevant table and cite the file (paths above are stable).
2. Prefer a single **named constant** in code (e.g. `config.PerSiteTimeout`) and reference that name here.
3. Distinguish **per-request colly policy** (no retry) from **BinderPOS multi-strategy fallback** (up to three different strategies, one try each).
