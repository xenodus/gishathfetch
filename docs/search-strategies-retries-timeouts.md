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
| Dedicated proxy per store search | 1 lease | `searchShop` in `api/controller/search.go` + `WithRequestDedicatedProxy` in `api/gateway/request_dedicated_proxy.go` | When dedicated proxies are configured, each store search acquires **one** dedicated-proxy lease for its own goroutine. Up to six concurrent store searches (see `maxConcurrentStoreSearches`) share the worker pool, but at most **three** proxy-backed searches may hold a dedicated lease at once (`DedicatedProxySearchMaxConcurrent` in `api/gateway/dedicated_proxy_search_gate.go`). Additional proxy-backed stores wait for a slot before leasing. |

---

## Backend: domain request pacing (all colly + `WaitForDomainRequestSlot` users)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Minimum interval between requests to the **same host** | 200ms | `domainRequestMinInterval` in `api/gateway/domain_rate_limiter.go` | Per reservation: `reservedUntil = nextAllowed + minInterval`. The first request for a host is immediate; later requests wait until the prior reservation expires. If the wait is cancelled, the limiter can roll back that reservation. |

## Backend: dedicated proxy env (`api/gateway/util/dedicated_proxy.go`)

| Item | Value | Notes |
|------|--------|--------|
| Configured slots | **`DEDICATED_PROXY_1`** … **`DEDICATED_PROXY_7`** | Each value is `host\|port\|username\|password` (pipe-separated). Empty or incomplete entries are ignored when building URLs. |
| Dedicated proxy search concurrency | 3 in-flight | `DedicatedProxySearchMaxConcurrent` in `api/gateway/dedicated_proxy_search_gate.go` | Caps how many store searches may hold a dedicated-proxy lease at once so datacenter egress does not burst every configured slot. |
| Dynamic fallback proxy | **`DYNAMIC_PROXY`** | Uses the same `host\|port\|username\|password` format as `DEDICATED_PROXY_*` (full proxy URLs are also accepted). BinderPOS reserves it for the final two fallback attempts (`scrap-dynamic` and `decklist-dynamic`) after dedicated and direct/no-proxy scrap and decklist tries. Disabled by default via `USE_DYNAMIC_PROXY`. |
| Dynamic proxy concurrency | 3 in-flight | `dynamicProxyMaxConcurrent` in `api/gateway/dynamic_proxy_gate.go` | Caps concurrent requests through `DYNAMIC_PROXY` across all stores and strategies so the final BinderPOS fallbacks do not burst the proxy endpoint and trigger 429. |
| Residential proxy | **`RESIDENTIAL_PROXY_1`** | Optional residential proxy for stores that rate-limit datacenter IPs (currently 5 Mana). Uses the same `host\|port\|username\|password` format. |
| Dynamic proxy toggle | **`USE_DYNAMIC_PROXY`** | When `true`, dynamic proxy fallback is enabled if `DYNAMIC_PROXY` is set. Defaults to **disabled** when unset or invalid. |

---

## Live BinderPOS integration tests

Some tests in `api/gateway/binderpos/*_test.go` hit real stores and proxies. They run only when **`RUN_BINDERPOS_LIVE_TESTS=1`** is set (default `make test` skips them to avoid rate limits and flaky remote dependencies).

---

## Backend: BinderPOS (storefront scraper and decklist fallbacks)

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| BinderPOS stores **with** Storefront access token | **graphql-dedicated** → **graphql-direct** → **scrap-dedicated** → **scrap-direct** → **decklist-dedicated** → **decklist-direct** → **scrap-dynamic** → **decklist-dynamic** | **5s** per step: `binderposAttemptTimeout` (`config.SearchAttemptTimeout`) in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. GraphQL uses dedicated then direct only (no dynamic GraphQL). Dynamic proxy remains reserved for the final scrap/decklist attempts. |
| BinderPOS stores **without** token | **scrap-dedicated** → **scrap-direct** → **decklist-dedicated** → **decklist-direct** → **scrap-dynamic** → **decklist-dynamic** | Same as above without the GraphQL steps. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| **scrap-dedicated** proxy selection (colly) | Request-scoped lease per store search when dedicated proxies are configured; otherwise random dedicated | `searchShop` + `selectOutboundProxy` in `api/gateway/collector.go` | Each store search holds one dedicated-proxy lease and pins it on its context. All `scrap-dedicated` attempts for that store reuse that URL. When `UseLeasedDedicatedProxy` is **true**, per-collector leases apply only when no request-scoped proxy is set. |
| Colly for BinderPOS scrapes | 5s | `SetRequestTimeout(binderposAttemptTimeout)` in `api/gateway/binderpos/scrap.go` | Same as `config.SearchAttemptTimeout`. |
| Decklist portal concurrency | 4 in-flight | `binderposPortalMaxConcurrent` in `api/gateway/binderpos/storefront_portal_gate.go` | Caps concurrent requests to `portal.binderpos.com` across stores in one search. |
| Decklist transient retries | Up to 3 sends | `binderposDecklistMaxAttempts` in `api/gateway/binderpos/storefront.go`; `doDecklistRequestWithRetry` in `storefront_decklist_retry.go` | Retries 429/5xx and network errors with equal-jitter backoff (300ms base, 2.5s cap) and honors `Retry-After`. |
| “Retries” | N/A (sequential fallbacks) | `runFallbackAttempts` in `storefront_fallback.go` | Stops on the first attempt that returns **cards**. An empty **GraphQL** or **scrape** attempt without error is **final** and later strategies are not tried. An empty **decklist** attempt skips remaining decklist strategies. HTTP **5xx** on scrape is **final**. GraphQL errors (including 5xx) fall through to HTML scrap. Returns the last annotated error if all attempts fail. This is **not** exponential backoff retry of a single scrape request. |
| Storefront GraphQL | Public per-store `accessToken` | `api/gateway/binderpos/storefront_graphql.go`; tokens in each store package (`StoreStorefrontAccessToken`) | Shopify Storefront `search` with `available: true`. MTG filtered client-side via product type/tags. Variant deep-links include `?variant=`. Enabled only when the store configures a token. |
| scrap-dynamic 429 retries | Up to 3 sends | `scrapDynamicMaxAttempts` in `api/gateway/binderpos/scrap_dynamic.go` | Retries 429 responses with equal-jitter backoff (reuses decklist backoff constants) and a fresh collector so the rotating proxy egresses from a new IP. Honors the per-attempt timeout budget. |

---

## Backend: non-BinderPOS stores (Agora, Cards Central, Dueller's Point, 5 Mana, Mox & Lotus, Cards & Collections, TCG Marketplace)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Outbound proxy policy | Random dedicated → dynamic → direct | `selectOutboundProxy` in `api/gateway/collector.go` | Same single-attempt policy as default optimized colly collectors. When a store search holds a request-scoped dedicated lease, colly and `DoOutboundGET` reuse that URL. When no lease is pinned and `DEDICATED_PROXY_*` is configured, each outbound store falls back to one random dedicated proxy. |
| `net/http` scrapers / APIs | Direct → one random dedicated proxy → dynamic fallback | `DoOutboundGET` / `DoOutboundRoundTrip` in `api/gateway/outbound_get.go` | Used by Agora, Dueller's Point, Mox & Lotus, Cards & Collections, and TCG Marketplace. Reuses the per-store dedicated lease when set. Each transport is tried once per store (one dedicated slot, not every configured proxy). Client errors (4xx) and connection errors advance immediately to the next transport. |
| 5 Mana Shopify search | **graphql** → **html** (section scrape); Residential → dedicated → dynamic | `api/gateway/fivemana/search.go`, `graphql.go` | Primary: Storefront GraphQL `search` with public access token (variants include condition/price). Fallback: Dawn `main-search` HTML section scrape on GraphQL error. Both paths skip direct requests and try `RESIDENTIAL_PROXY_1` first, then dedicated, then dynamic. Client errors (4xx) fail over immediately. |
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
