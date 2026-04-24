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

---

## Backend: BinderPOS (storefront + scraper fallbacks)

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| `shopifyDomain` **non-empty** (normal) | 1) **api-dedicated** → 2) **api-shared** → 3) **scrap-shared** (shared-proxy scrape) | **10s** per step: `binderposAttemptTimeout` in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. HTTP clients in `storefront_client.go` use the same `binderposAttemptTimeout`. |
| `shopifyDomain` **empty** | Scrape only (**scrap-shared** path, single `runWithAttemptTimeout`) | **10s** (same constant). No API attempts. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Colly for BinderPOS scrapes | 10s | `SetRequestTimeout(binderposAttemptTimeout)` in `api/gateway/binderpos/scrap.go` | Tighter than generic `PerSiteTimeout` (20s) for binderpos scrape collectors. |
| “Retries” | N/A (sequential fallbacks) | `searchWithFallback` | Stops on first **success**; returns last error if all three fail. This is **not** exponential backoff retry of a single request. |

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
