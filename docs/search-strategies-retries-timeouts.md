# Search strategies, retries, and timing

This document records **where** the app configures search behavior, **timeouts**, **fallback/attempt ordering**, **concurrency limits**, and **request pacing**. It is meant for code agents and maintainers: when you change a constant, update this file in the same PR.

---

## Backend: controller (multi-store search)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Per-store deadline | 16s | `config.PerSiteTimeout` in `api/pkg/config/config.go`; used in `searchShop` as `context.WithTimeout` in `api/controller/search.go` | One goroutine per selected store; each `LGS.Search` runs under this cap. |
| Per-attempt timeout (default) | 5s | `config.SearchAttemptTimeout` in `api/pkg/config/config.go` | Bounds each BinderPOS strategy step and default colly scrape request. |
| Agora per-attempt timeout | 10s | `config.AgoraSearchAttemptTimeout` in `api/pkg/config/config.go`; applied in `api/gateway/agora/search.go` | Agora keeps a longer single-scrape cap. |
| Colly request timeout (default scrapers) | 5s | `applyCollectorDefaults` → `c.SetRequestTimeout(config.SearchAttemptTimeout)` in `api/gateway/collector.go` | Overrides gocolly’s default 10s for optimized collectors. |
| Minimum end-to-end response time | 1s | `responseThreshold` in `searchShops` in `api/controller/search.go` | If all stores finish in under 1s, the handler **sleeps** the remainder so the API “feels” less instant. |
| Colly HTTP retries | None | `api/gateway/collector.go` (`configureRequestOptimizations`, `registerNoRetryErrorHandler`) | **Single HTTP attempt** per colly request path; no automatic colly/gateway retry of failed visits. |
| BinderPOS store concurrency | 12 | `binderposMaxConcurrent` in `api/controller/search.go` | Semaphore limits how many binderpos-backed shops run at once. Non-binderpos stores are not limited by this gate. |
| BinderPOS shared portal-host concurrency | 4 | `binderposPortalMaxConcurrent` in `api/gateway/binderpos/storefront_portal_gate.go` | Separate semaphore limiting concurrent decklist requests to the single shared host `portal.binderpos.com`. Every binderpos store's primary lookup funnels into this one host, so this caps the per-host burst independently of the per-store gate above. Acquired in `searchByBinderposDecklistAPI`; respects context cancellation. |

---

## Backend: domain request pacing (all colly + `WaitForDomainRequestSlot` users)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Minimum interval between requests to the **same host** | 200ms | `domainRequestMinInterval` in `api/gateway/domain_rate_limiter.go` | Per reservation: `reservedUntil = nextAllowed + minInterval`. The first request for a host is immediate; later requests wait until the prior reservation expires. If the wait is cancelled, the limiter can roll back that reservation. |
| Always-paced (shared) hosts | `portal.binderpos.com` | `RegisterAlwaysPacedDomain` / `alwaysPacedDomains` in `api/gateway/domain_rate_limiter.go`; registered in `api/gateway/binderpos/storefront_portal_gate.go` `init` | Hosts in this set stay paced **even when a caller opts out** via `WithDomainRequestPacingDisabled`. The opt-out is meant for per-store hosts (a store's first attempt), where skipping the inter-request delay is safe; on a shared host it would let concurrent stores burst the same upstream and cause 429/503. |

## Backend: dedicated proxy env (`api/gateway/util/dedicated_proxy.go`)

| Item | Value | Notes |
|------|--------|--------|
| Configured slots | **`DEDICATED_PROXY_1`** … **`DEDICATED_PROXY_7`** | Each value is `host\|port\|username\|password` (pipe-separated). Empty or incomplete entries are ignored when building URLs. |
| Dynamic fallback proxy | **`DYNAMIC_PROXY`** | Uses the same `host\|port\|username\|password` format as `DEDICATED_PROXY_*` (full proxy URLs are also accepted). BinderPOS reserves it for the final fallback after dedicated and direct/no-proxy attempts. |
| Dynamic proxy toggle | **`USE_DYNAMIC_PROXY`** | When `false`, dynamic proxy fallback is disabled even if `DYNAMIC_PROXY` is set. Defaults to enabled when unset or invalid. |

---

## Live BinderPOS integration tests

Some tests in `api/gateway/binderpos/*_test.go` hit real stores and proxies. They run only when **`RUN_BINDERPOS_LIVE_TESTS=1`** is set (default `make test` skips them to avoid rate limits and flaky remote dependencies).

---

## Backend: BinderPOS (storefront + scraper fallbacks)

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| `shopifyDomain` **non-empty** (normal) | 1) **api-dedicated** → 2) **api-direct** → 3) **scrap-dedicated** → 4) **scrap-direct** → 5) **api-dynamic** → 6) **scrap-dynamic** | **5s** per step: `binderposAttemptTimeout` (`config.SearchAttemptTimeout`) in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. HTTP clients in `storefront_client.go` use the same constant. |
| `shopifyDomain` **empty** or **`ScrapOnly`** | **scrap-dedicated** → **scrap-direct** → **scrap-dynamic** (three `runWithAttemptTimeout` steps) | **5s** per step (same constant). No API attempts. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| **api-dedicated** proxy selection (storefront HTTP client) | Round-robin | `nextBinderposStorefrontProxyURL` + `binderposDedicatedProxySeq` in `api/gateway/binderpos/storefront_client.go` | When `UseLeasedDedicatedProxy` is **false** (default in `api/pkg/config/config.go`), each storefront API call cycles **proxy₁ → … → proxyₙ** then repeats, using all URLs from `GetDedicatedProxyURLs()`. When `UseLeasedDedicatedProxy` is **true**, selection is unchanged: `LeaseDedicatedProxyURL` from the dedicated pool. |
| **api-dynamic** proxy selection (storefront HTTP client) | Fixed env URL | `searchByStorefrontAPIDynamic` in `api/gateway/binderpos/storefront_client.go` | Uses `DYNAMIC_PROXY` as the authenticated proxy URL for the final BinderPOS API fallback attempt. |
| Colly for BinderPOS scrapes | 5s | `SetRequestTimeout(binderposAttemptTimeout)` in `api/gateway/binderpos/scrap.go` | Same as `config.SearchAttemptTimeout`. |
| “Retries” | N/A (sequential fallbacks) | `searchWithFallback` | Stops on first **success**; returns last error if all attempts fail. This is **not** exponential backoff retry of a single request. |

---

## Backend: non-BinderPOS stores (Agora, Cards Central, Dueller's Point, 5 Mana, Mox & Lotus, Cards & Collections, TCG Marketplace)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Outbound proxy policy | Random dedicated → dynamic → direct | `selectOutboundProxy` in `api/gateway/collector.go` | Same single-attempt policy as default optimized colly collectors. When `DEDICATED_PROXY_*` is configured, each search picks one dedicated proxy uniformly at random. |
| `net/http` scrapers / APIs | Direct → dedicated proxies → dynamic fallback | `DoOutboundGET` / `DoOutboundRoundTrip` in `api/gateway/outbound_get.go` | Used by Agora, Dueller's Point, 5 Mana, Mox & Lotus, Cards & Collections, and TCG Marketplace. Each transport is tried once per search; timeouts, connection errors, 403, and 429 advance to the next transport. |
| Cards Central API | Direct only | `http.Client` in `api/gateway/cardscentral/search.go` | Always uses a direct client; does not route through dedicated or dynamic proxies. |

---

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
