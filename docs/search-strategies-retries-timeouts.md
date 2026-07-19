# Search strategies, retries, and timing

This document records **where** the app configures search behavior, **timeouts**, **fallback/attempt ordering**, **concurrency limits**, and **request pacing**. It is meant for code agents and maintainers: when you change a constant, update this file in the same PR.

---

## Backend: controller (multi-store search)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Per-store deadline | 20s | `config.PerSiteTimeout` in `api/pkg/config/config.go`; used in `searchShop` as `context.WithTimeout` in `api/controller/search.go` | One goroutine per selected store; each `LGS.Search` runs under this cap. |
| Per-attempt timeout (default) | 5s | `config.SearchAttemptTimeout` in `api/pkg/config/config.go` | Bounds each BinderPOS strategy step, default colly scrape request, and most `DoOutboundGET` / `DoOutboundRoundTrip` calls. |
| Agora per-attempt timeout | 10s | `config.AgoraSearchAttemptTimeout` in `api/pkg/config/config.go`; applied in `api/gateway/agora/search.go` | Agora keeps a longer single-scrape cap. |
| Colly request timeout (default scrapers) | 5s | `applyCollectorDefaults` → `c.SetRequestTimeout(config.SearchAttemptTimeout)` in `api/gateway/collector.go` | Overrides gocolly’s default 10s for optimized collectors. |
| Max concurrent store searches | 6 | `maxConcurrentStoreSearches` in `api/controller/search.go` | Worker pool size when fanning out to selected stores. |
| Minimum end-to-end response time | 1s | `responseThreshold` in `searchShops` in `api/controller/search.go` | If all stores finish in under 1s, the handler **sleeps** the remainder so the API “feels” less instant. |
| Colly HTTP retries | None | `api/gateway/collector.go` (`configureRequestOptimizations`, `registerNoRetryErrorHandler`) | **Single HTTP attempt** per colly request path; no automatic colly/gateway retry of failed visits. |
| Dedicated proxy per store search | 1 lease | `searchShop` in `api/controller/search.go` + `WithRequestDedicatedProxy` in `api/gateway/request_dedicated_proxy.go` | When dedicated proxies are configured, each store search acquires **one** dedicated-proxy lease for its own goroutine. Up to six concurrent store searches share the worker pool, but at most **three** proxy-backed searches may hold a dedicated lease at once (`DedicatedProxySearchMaxConcurrent` in `api/gateway/dedicated_proxy_search_gate.go`). Additional proxy-backed stores wait for a slot before leasing. |

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
| Residential proxy | **`RESIDENTIAL_PROXY_1`** | Optional residential proxy for stores that block datacenter IPs behind Cloudflare. Uses the same `host\|port\|username\|password` format. |
| Dynamic proxy toggle | **`USE_DYNAMIC_PROXY`** | When `true`, dynamic proxy fallback is enabled if `DYNAMIC_PROXY` is set. Defaults to **disabled** when unset or invalid. |

---

## Live BinderPOS integration tests

Some tests in `api/gateway/binderpos/*_test.go` hit real stores and proxies. They run only when **`RUN_BINDERPOS_LIVE_TESTS=1`** is set (default `make test` skips them to avoid rate limits and flaky remote dependencies).

---

## Backend: BinderPOS (storefront scraper and decklist fallbacks)

For **field-level feature parity** between HTML scrape, Storefront GraphQL, Decklist API, scrap variants, and transport modes, see [`binderpos-search-feature-parity.md`](binderpos-search-feature-parity.md).

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| BinderPOS stores **with** Storefront access token | **graphql-dedicated** → **graphql-direct** → **scrap-dedicated** → **scrap-direct** → **decklist-dedicated** → **decklist-direct** → **scrap-dynamic** → **decklist-dynamic** | **5s** per step: `binderposAttemptTimeout` (`config.SearchAttemptTimeout`) in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. GraphQL uses dedicated then direct only (no dynamic GraphQL). Dynamic proxy remains reserved for the final scrap/decklist attempts. |
| BinderPOS stores **without** token | **scrap-dedicated** → **scrap-direct** → **decklist-dedicated** → **decklist-direct** → **scrap-dynamic** → **decklist-dynamic** | Same as above without the GraphQL steps. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Colly proxy selection (scrap steps) | Request-scoped dedicated → per-collector lease → random dedicated → dynamic → direct | `selectOutboundProxy` in `api/gateway/collector.go` | Each colly collector makes **one** outbound attempt using the first available mode. When `searchShop` pins a request-scoped dedicated lease, scrap steps reuse that URL. When `UseLeasedDedicatedProxy` is **true**, per-collector leases apply only when no request-scoped proxy is set. |
| Colly for BinderPOS scrapes | 5s | `SetRequestTimeout(binderposAttemptTimeout)` in `api/gateway/binderpos/scrap.go` | Same as `config.SearchAttemptTimeout`. |
| Decklist portal concurrency | 4 in-flight | `binderposPortalMaxConcurrent` in `api/gateway/binderpos/storefront_portal_gate.go` | Caps concurrent requests to `portal.binderpos.com` across stores in one search. |
| Decklist requests | Single send | `doDecklistRequestWithRetry` in `api/gateway/binderpos/storefront_decklist_retry.go` | No automatic retries on 429/5xx or network errors. |
| “Retries” | N/A (sequential fallbacks) | `runFallbackAttempts` in `storefront_fallback.go` | Stops on the first attempt that returns **cards**. An empty **GraphQL** or **scrape** attempt without error is **final** and later strategies are not tried. An empty **decklist** attempt skips remaining decklist strategies. HTTP **5xx** on scrape or GraphQL is **final**. Other GraphQL errors fall through to HTML scrap. Returns the last annotated error if all attempts fail. This is **not** exponential backoff retry of a single scrape request. |
| Storefront GraphQL | Public per-store `accessToken` | `api/gateway/binderpos/storefront_graphql.go`; tokens in each store package (`StoreStorefrontAccessToken`) | Shopify Storefront `search` with `available: true`. MTG filtered client-side via product type/tags. Variant deep-links include `?variant=`. Enabled only when the store configures a token. |
| scrap-dynamic | Single send | `scrapDynamic` in `api/gateway/binderpos/scrap_dynamic.go` | No automatic 429 retries; each attempt uses one dynamic-proxy collector. |

### BinderPOS stores (registry in `api/controller/search.go`)

| Store | GraphQL token | HTML scrap variant | Max strategy steps |
|-------|---------------|--------------------|--------------------|
| Card Affinity | No | 2 | 6 |
| Cards Citadel | Yes | 1 | 8 |
| Flagship | Yes | 2 | 8 |
| Fyendal Hobby | Yes | 4 | 8 |
| Games Haven | Yes | 3 | 8 |
| GOG | Yes | 3 | 8 |
| Hideout | Yes | 3 | 8 |
| Hideyoshi | Yes | 2 | 8 |
| Mana Pro | Yes | 2 | 8 |
| MTG Asia | Yes | 2 | 8 |
| One MTG | Yes | 2 | 8 |

All BinderPOS stores configure a Shopify domain for decklist steps. Card Affinity is the only BinderPOS store without a Storefront GraphQL token.

---

## Backend: non-BinderPOS stores

Shared `net/http` transport fallback for `DoOutboundGET` / `DoOutboundRoundTrip` (`api/gateway/outbound_get.go`): **direct → dedicated (request-scoped lease or one random slot) → dynamic**. Each transport is tried once; client errors (4xx) and connection errors advance immediately to the next transport. No automatic retry of the same transport.

| Store | Strategy | Per-attempt timeout | Proxy / transport order | Retries |
|-------|----------|----------------------|-------------------------|---------|
| Agora Hobby | HTML search page (`/store/search`) | 10s | **Dedicated → dynamic**; skips direct on `agorahobby.com` (browser TLS via `SkipWebBotAuth` + `BROWSER_TLS_EMULATION_ENABLED`) | Transport fallback only |
| 5 Mana | **graphql** → **html** (Dawn `main-search` section) | 5s per path | **Dedicated → dynamic**; skips direct on `5-mana.sg` | GraphQL 5xx is final; other GraphQL errors fall through to HTML. Transport fallback per path. |
| Cards Central | JSON API (`/api/lgs/search?q=…`) | 5s | Direct → dedicated → dynamic | Transport fallback only |
| Dueller's Point | HTML search page (`/products/search`) | 5s | Direct → dedicated → dynamic | Transport fallback only |
| Mox & Lotus | JSON API GET (`/api/products?search=…`) | 5s | Direct → dedicated → dynamic | Transport fallback only |
| Cards & Collections | Elasticsearch-style POST (`/api/catalog/`) | 5s | Direct → dedicated → dynamic | Transport fallback only |
| The TCG Marketplace | CardLink POST (`:3501/encoder/advancedsearch`) | 5s | Direct → dedicated → dynamic | Transport fallback only |

Store implementations: `api/gateway/agora/search.go`, `api/gateway/fivemana/search.go` + `graphql.go`, `api/gateway/cardscentral/search.go`, `api/gateway/duellerpoint/search.go`, `api/gateway/moxandlotus/search.go`, `api/gateway/cardsandcollection/search.go`, `api/gateway/tcgmarketplace/search.go`.

For Agora and 5 Mana, `SkipDirect` is cleared when the host is not the production domain so httptest unit tests can use the direct transport.

---

## Backend: CK price refresh (`api/gateway/cardkingdom/`)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| CK pricelist HTTP timeout | 12m | `ckPricelistHTTPTimeout` in `api/gateway/cardkingdom/pricelist_fetch.go` | Bounds the full `DoOutboundGET` round trip, including streaming the ~65MB JSON body through `CK_PRICELIST_PROXY`. |
| CK pricelist fetch timeout | 13m | `ckPricelistFetchTimeout` in `api/gateway/cardkingdom/pricelist_fetch.go` | Context deadline for download + JSON decode + cheapest-listing aggregation. |
| CK pricelist body-read progress logs | every 15s | `ckPricelistBodyReadLogInterval` in `api/gateway/cardkingdom/pricelist_fetch.go` | Emits bytes read (and `%` when `Content-Length` is present) while the body streams. |

---

## Frontend

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
2. Prefer a single **named constant** in code (e.g. `config.PerSiteTimeout`, `config.SearchAttemptTimeout`) and reference that name here. When a store hardcodes a timeout, document the literal and source file.
3. Distinguish **per-request colly policy** (no retry) from **BinderPOS multi-strategy fallback** (up to **eight** strategies when GraphQL token and decklist are configured, **six** without GraphQL; one try each per strategy step).
4. When adding a store, update the per-store tables in the BinderPOS or non-BinderPOS section and register it in `api/controller/search.go`.
