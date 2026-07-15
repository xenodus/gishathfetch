# Search strategies, retries, and timing

This document records **where** the app configures search behavior, **timeouts**, **fallback/attempt ordering**, **concurrency limits**, and **request pacing**. It is meant for code agents and maintainers: when you change a constant, update this file in the same PR.

---

## Backend: controller (multi-store search)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Per-store deadline | 20s | `config.PerSiteTimeout` in `api/pkg/config/config.go`; used in `searchShop` as `context.WithTimeout` in `api/controller/search.go` | One goroutine per selected store; each `LGS.Search` runs under this cap. |
| Per-attempt timeout (default) | 5s | `config.SearchAttemptTimeout` in `api/pkg/config/config.go` | Bounds each BinderPOS strategy step and default colly scrape request. |
| Agora per-attempt timeout | 10s | `config.AgoraSearchAttemptTimeout` in `api/pkg/config/config.go`; applied in `api/gateway/agora/search.go` | Agora keeps a longer single-scrape cap. |
| Colly request timeout (default scrapers) | 5s | `applyCollectorDefaults` → `c.SetRequestTimeout(config.SearchAttemptTimeout)` in `api/gateway/collector.go` | Overrides gocolly’s default 10s for optimized collectors. |
| Minimum end-to-end response time | 1s | `responseThreshold` in `searchShops` in `api/controller/search.go` | If all stores finish in under 1s, the handler **sleeps** the remainder so the API “feels” less instant. |
| Colly HTTP retries | None | `api/gateway/collector.go` (`configureRequestOptimizations`, `registerNoRetryErrorHandler`) | **Single HTTP attempt** per colly request path; no automatic colly/gateway retry of failed visits. |
| Dedicated proxy per search request | 1 lease | `fetchCardsConcurrently` in `api/controller/search.go` + `WithRequestDedicatedProxy` in `api/gateway/request_dedicated_proxy.go` | When dedicated proxies are configured, the controller acquires **one** dedicated-proxy lease for the whole search. All stores in that search share the pinned URL. Concurrent searches take distinct slots from the shared pool (up to seven); an eighth concurrent search waits until a slot is released. |

---

## Backend: domain request pacing (all colly + `WaitForDomainRequestSlot` users)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Minimum interval between requests to the **same host** | 200ms | `domainRequestMinInterval` in `api/gateway/domain_rate_limiter.go` | Per reservation: `reservedUntil = nextAllowed + minInterval`. The first request for a host is immediate; later requests wait until the prior reservation expires. If the wait is cancelled, the limiter can roll back that reservation. |

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

## Backend: BinderPOS (storefront scraper and decklist fallbacks)

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| All BinderPOS stores | **scrap-dedicated** → **scrap-direct** → **scrap-dynamic** → **decklist-dedicated** → **decklist-direct** → **decklist-dynamic** (six `runWithAttemptTimeout` steps when a Shopify domain is configured) | **5s** per step: `binderposAttemptTimeout` (`config.SearchAttemptTimeout`) in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| **scrap-dedicated** proxy selection (colly) | Request-scoped lease when dedicated proxies are configured; otherwise random dedicated | `fetchCardsConcurrently` + `selectOutboundProxy` in `api/gateway/collector.go` | When dedicated proxies are configured, the controller holds one dedicated-proxy lease per search and pins it on the search context. All `scrap-dedicated` attempts for stores in that search reuse that URL. When `UseLeasedDedicatedProxy` is **true**, per-collector leases apply only when no request-scoped proxy is set. |
| Colly for BinderPOS scrapes | 5s | `SetRequestTimeout(binderposAttemptTimeout)` in `api/gateway/binderpos/scrap.go` | Same as `config.SearchAttemptTimeout`. |
| Decklist portal concurrency | 4 in-flight | `binderposPortalMaxConcurrent` in `api/gateway/binderpos/storefront_portal_gate.go` | Caps concurrent requests to `portal.binderpos.com` across stores in one search. |
| Decklist transient retries | Up to 3 sends | `binderposDecklistMaxAttempts` in `api/gateway/binderpos/storefront.go`; `doDecklistRequestWithRetry` in `storefront_decklist_retry.go` | Retries 429/5xx and network errors with equal-jitter backoff (300ms base, 2.5s cap) and honors `Retry-After`. |
| “Retries” | N/A (sequential fallbacks) | `runFallbackAttempts` in `storefront_fallback.go` | Stops on the first attempt that returns **cards**. An empty **scrape** attempt without error is **final** and decklist is not tried. An empty **decklist** attempt skips remaining decklist strategies. HTTP **5xx** on scrape is **final**. Returns the last annotated error if all attempts fail. This is **not** exponential backoff retry of a single scrape request. |

---

## Backend: non-BinderPOS stores (Agora, Cards Central, Dueller's Point, 5 Mana, Mox & Lotus, Cards & Collections, TCG Marketplace)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Outbound proxy policy | Random dedicated → dynamic → direct | `selectOutboundProxy` in `api/gateway/collector.go` | Same single-attempt policy as default optimized colly collectors. When a search holds a request-scoped dedicated lease, colly and `DoOutboundGET` reuse that URL. When no lease is pinned and `DEDICATED_PROXY_*` is configured, each outbound store falls back to one random dedicated proxy. |
| `net/http` scrapers / APIs | Direct → one random dedicated proxy → dynamic fallback | `DoOutboundGET` / `DoOutboundRoundTrip` in `api/gateway/outbound_get.go` | Used by Agora, Dueller's Point, 5 Mana, Mox & Lotus, Cards & Collections, and TCG Marketplace. Reuses the request-scoped dedicated lease when set. Each transport is tried once per store (one dedicated slot, not every configured proxy). 429 responses retry with backoff on the same transport before failing over; 403 and connection errors advance immediately. |
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
3. Distinguish **per-request colly policy** (no retry) from **BinderPOS multi-strategy fallback** (up to six different strategies when decklist is configured, one try each per strategy step).
