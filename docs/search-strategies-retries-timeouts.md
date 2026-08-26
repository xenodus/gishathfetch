# Search strategies, retries, and timing

This document records **where** the app configures search behavior, **timeouts**, **fallback/attempt ordering**, **concurrency limits**, and **request pacing**. It is meant for code agents and maintainers: when you change a constant, update this file in the same PR.

---

## Backend: controller (multi-store search)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Per-store deadline | 20s | `config.PerSiteTimeout` in `api/pkg/config/config.go`; used in `searchShop` as `context.WithTimeout` in `api/controller/search.go` | One goroutine per selected store; each `LGS.Search` runs under this cap. When the query contains diacritics, the original and ASCII-stripped forms are searched **in parallel** under the same per-store deadline, then merged and deduped. |
| Per-attempt timeout (direct) | 3s | `config.DirectSearchAttemptTimeout` in `api/pkg/config/config.go` | Bounds each direct-egress search attempt across BinderPOS steps, colly scrapes, and `DoOutboundGET` / `DoOutboundRoundTrip` direct transports. |
| Per-attempt timeout (dedicated) | 5s | `config.DedicatedSearchAttemptTimeout` in `api/pkg/config/config.go` | Bounds each dedicated-proxy search attempt across BinderPOS steps, colly scrapes, and `DoOutboundGET` / `DoOutboundRoundTrip` dedicated transports. |
| Agora per-attempt timeout | 20s | `config.AgoraSearchAttemptTimeout` in `api/gateway/agora/search.go` via `DirectAttemptTimeout` / `DedicatedAttemptTimeout` | Same as per-store deadline (`PerSiteTimeout`). |
| Mox & Lotus per-attempt timeout | 10s | `config.MoxAndLotusSearchAttemptTimeout` in `api/gateway/moxandlotus/search.go` via `DirectAttemptTimeout` / `DedicatedAttemptTimeout` | Longer than the default direct/dedicated attempt timeouts. |
| Colly request timeout (default scrapers) | 3s direct / 5s dedicated | `applyCollectorDefaults` and `applyCollectorHTTPClient` in `api/gateway/collector.go` | Direct collectors use `DirectSearchAttemptTimeout`; proxy-backed collectors use `DedicatedSearchAttemptTimeout`. BinderPOS scrap steps override explicitly in `scrap.go`. |
| Max concurrent store searches | 5 | `maxConcurrentStoreSearches` in `api/controller/search.go` | Worker pool size when fanning out to selected stores. |
| Minimum end-to-end response time | 1s | `responseThreshold` in `searchShops` in `api/controller/search.go` | If all stores finish in under 1s, the handler **sleeps** the remainder so the API “feels” less instant. |
| Card Kingdom enrichment on `/search` | parallel, **2s** cap | See [CK price on search](#backend-ck-price-on-search-apihandlersearchgo-and-refresh-apigatewaycardkingdom) below | Store fan-out and CK lookup run together; CK cannot delay the response past its timeout. |
| Colly HTTP retries | None | `api/gateway/collector.go` (`configureRequestOptimizations`, `registerNoRetryErrorHandler`) | **Single HTTP attempt** per colly request path; no automatic colly/gateway retry of failed visits. |
| Dedicated proxy per store search | 1 lease | `searchShop` in `api/controller/search.go` + `WithRequestDedicatedProxy` in `api/gateway/request_dedicated_proxy.go` | When dedicated proxies are configured, each store search acquires **one** dedicated-proxy lease for its own goroutine. Up to five concurrent store searches share the worker pool, but at most **three** proxy-backed searches may hold a dedicated lease at once (`DedicatedProxySearchMaxConcurrent` in `api/gateway/dedicated_proxy_search_gate.go`). Additional proxy-backed stores wait for a slot before leasing. |

---

## Backend: domain request pacing (all colly + `WaitForDomainRequestSlot` users)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Minimum interval between requests to the **same host** | 200ms | `domainRequestMinInterval` in `api/gateway/domain_rate_limiter.go` | Per reservation: `reservedUntil = nextAllowed + minInterval`. The first request for a host is immediate; later requests wait until the prior reservation expires. If the wait is cancelled, the limiter can roll back that reservation. |

## Backend: dedicated proxy env (`api/gateway/util/dedicated_proxy.go`)

| Item | Value | Notes |
|------|--------|--------|
| Configured slots | **`DEDICATED_PROXY_1`** … **`DEDICATED_PROXY_7`** | Each value is `host\|port\|username\|password` (pipe-separated). Empty or incomplete entries are ignored when building URLs. |
| Dedicated proxy toggle | **`USE_DEDICATED_PROXY`** | When `false`, dedicated proxy transports are skipped even if `DEDICATED_PROXY_*` are set. Defaults to **enabled** when unset or invalid. |
| Dedicated proxy search concurrency | 3 in-flight | `DedicatedProxySearchMaxConcurrent` in `api/gateway/dedicated_proxy_search_gate.go` | Caps how many store searches may hold a dedicated-proxy lease at once so datacenter egress does not burst every configured slot. |
| Residential proxy | **`RESIDENTIAL_PROXY_1`** | Optional residential proxy for stores that block datacenter IPs behind Cloudflare. Uses the same `host\|port\|username\|password` format. |

---

## Live BinderPOS integration tests

Some tests in `api/gateway/binderpos/*_test.go` hit real stores and proxies. They run only when **`RUN_BINDERPOS_LIVE_TESTS=1`** is set (default `make test` skips them to avoid rate limits and flaky remote dependencies).

---

## Backend: BinderPOS (storefront scraper fallbacks)

For **field-level feature parity** between HTML scrape, Storefront GraphQL, Decklist API, scrap variants, and transport modes, see [`binderpos-search-feature-parity.md`](binderpos-search-feature-parity.md).

`api/gateway/binderpos/storefront_fallback.go` and `api/gateway/binderpos/storefront_search.go` define a **sequential multi-strategy** flow (not the same as colly “retry n times on failure” for one URL).

| Scenario | Order of strategies (each step is one attempt) | Per-step attempt timeout / HTTP client |
|----------|--------------------------------------------------|----------------------------------------|
| BinderPOS stores **with** Storefront access token | **graphql-direct** → **graphql-dedicated** → **scrap-direct** → **scrap-dedicated** | **3s** direct / **5s** dedicated per step: `binderposDirectAttemptTimeout` / `binderposDedicatedAttemptTimeout` in `api/gateway/binderpos/storefront.go`; `runWithAttemptTimeout` in `storefront_search.go`. |
| BinderPOS stores **without** token | **scrap-direct** → **scrap-dedicated** | Same as above without the GraphQL steps. |

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| Colly proxy selection (scrap steps) | Request-scoped dedicated → per-collector lease → random dedicated → direct | `selectOutboundProxy` in `api/gateway/collector.go` | Each colly collector makes **one** outbound attempt using the first available mode. When `searchShop` pins a request-scoped dedicated lease, scrap steps reuse that URL. When `UseLeasedDedicatedProxy` is **true**, per-collector leases apply only when no request-scoped proxy is set. |
| Colly for BinderPOS scrapes | 3s direct / 5s dedicated | `SetRequestTimeout` in `api/gateway/binderpos/scrap.go` | Direct scrap steps use `binderposDirectAttemptTimeout`; dedicated scrap steps use `binderposDedicatedAttemptTimeout`. |
| Decklist portal concurrency | 4 in-flight | `binderposPortalMaxConcurrent` in `api/gateway/binderpos/storefront_portal_gate.go` | Caps concurrent requests to `portal.binderpos.com` when decklist helpers are called directly (not used in the live search chain). |
| Decklist requests | Single send | `doDecklistRequestWithRetry` in `api/gateway/binderpos/storefront_decklist_retry.go` | No automatic retries on 429/5xx or network errors. Decklist is implemented but not wired into `storefront_search.go`. |
| “Retries” | N/A (sequential fallbacks) | `runFallbackAttempts` in `storefront_fallback.go` | Stops on the first attempt that returns **cards**. An empty **GraphQL** or **scrape** attempt without error is **final** and later strategies are not tried. HTTP **5xx** on scrape or GraphQL is **final**. Other GraphQL errors fall through to HTML scrap. Returns the last annotated error if all attempts fail. This is **not** exponential backoff retry of a single scrape request. |
| Storefront GraphQL | Public per-store `accessToken` | `api/gateway/binderpos/storefront_graphql.go`; tokens in each store package (`StoreStorefrontAccessToken`) | Shopify Storefront `search` with `available: true`. MTG filtered client-side via product type/tags. Variant deep-links include `?variant=`. Enabled only when the store configures a token. |

### BinderPOS stores (registry in `api/controller/search.go`)

| Store | GraphQL token | HTML scrap variant | Max strategy steps |
|-------|---------------|--------------------|--------------------|
| Arcane Sanctum | Yes | 2 | 4 |
| Card Affinity | No | 2 | 2 |
| Cards Citadel | Yes | 1 | 4 |
| Flagship | Yes | 2 | 4 |
| Fyendal Hobby | Yes | 4 | 4 |
| Games Haven | Yes | 3 | 4 |
| GOG | Yes | 3 | 4 |
| Hideout | Yes | 3 | 4 |
| Hideyoshi | Yes | 2 | 4 |
| Mana Pro | Yes | 2 | 4 |
| MTG Asia | Yes | 2 | 4 |
| One MTG | Yes | 2 | 4 |

Card Affinity is the only BinderPOS store without a Storefront GraphQL token.

---

## Backend: non-BinderPOS stores

Shared `net/http` transport fallback for `DoOutboundGET` / `DoOutboundRoundTrip` (`api/gateway/outbound_get.go`): **direct (3s) → dedicated (5s, request-scoped lease or one random slot)** when dedicated proxies are enabled. Callers can omit direct with `SkipDirect`. Each transport is tried once; client errors (4xx) and connection errors advance immediately to the next transport. No automatic retry of the same transport.

| Store | Strategy | Per-attempt timeout | Proxy / transport order | Retries |
|-------|----------|----------------------|-------------------------|---------|
| Agora Hobby | HTML search page (`/store/search`) | 20s | **Direct → dedicated** (browser TLS via `SkipWebBotAuth` + `BROWSER_TLS_EMULATION_ENABLED`) | Transport fallback only |
| 5 Mana | **graphql** → **html** (Dawn `main-search` section) | 3s direct / 5s dedicated per path | **Direct → dedicated** | GraphQL 5xx is final; other GraphQL errors fall through to HTML. Transport fallback per path. |
| Tefuda | **graphql** → **html** (Ride theme; MTG singles `product_type` filter) | 3s direct / 5s dedicated per path | **Direct → dedicated** | GraphQL 5xx is final; other GraphQL errors fall through to HTML. Transport fallback per path. |
| Cards Central | JSON API (`/api/lgs/search?q=…`) | 3s direct / 5s dedicated | Direct → dedicated | Transport fallback only |
| Dueller's Point | HTML search page (`/products/search`) | 3s direct / 5s dedicated | Direct → dedicated | Transport fallback only |
| Mox & Lotus | JSON API GET (`/api/products?search=…`, `limit=24`) | 10s | **Direct → dedicated**; browser JSON headers + `SkipWebBotAuth` | Transport fallback only |
| Cards & Collections | Elasticsearch-style POST (`/api/catalog/`) | 3s direct / 5s dedicated | Direct → dedicated | Transport fallback only |
| The TCG Marketplace | CardLink POST (`:3501/encoder/advancedsearch`) | 3s direct / 5s dedicated | Direct → dedicated → dynamic | Transport fallback only |

Store implementations: `api/gateway/agora/search.go`, `api/gateway/fivemana/search.go` + `graphql.go`, `api/gateway/tefuda/search.go` + `graphql.go`, `api/gateway/cardscentral/search.go`, `api/gateway/duellerpoint/search.go`, `api/gateway/moxandlotus/search.go`, `api/gateway/cardsandcollection/search.go`, `api/gateway/tcgmarketplace/search.go`.

All stores with dedicated proxy configured use **direct before dedicated** transport fallback when dedicated proxies are enabled.

---

## Backend: CK price on search (`api/handler/search.go`) and refresh (`api/gateway/cardkingdom/`)

| Item | Value | Source | Notes |
|------|--------|--------|--------|
| CK lookup on `/search` | parallel with store search | `Search` in `api/handler/search.go` | When enabled (`CKPriceLookupEnabled`), enrichment runs in a sibling goroutine; response waits for search **and** CK (subject to the timeout below). |
| CK lookup timeout on `/search` | 2s | `config.CKPriceLookupTimeout` | Cancels Scryfall + DynamoDB work so store results are not blocked by a slow enrichment path. Timed-out lookups omit `cardKingdomPrice`. |
| Scryfall verify HTTP timeout | 3s per request | `httpClientTimeout` in `api/gateway/scryfall/verify.go` | `VerifyCardName` may do autocomplete then exact-named; each call uses this client timeout. |
| CK pricelist transport | direct → residential | `downloadCKPricelist` in `api/gateway/cardkingdom/pricelist_fetch.go` | Tries direct egress first; on failure retries via `RESIDENTIAL_PROXY_1`, then `CK_PRICELIST_PROXY` when configured. |
| CK pricelist HTTP timeout | 13m | `ckPricelistHTTPTimeout` in `api/gateway/cardkingdom/pricelist_fetch.go` | Bounds the full `DoOutboundGET` round trip, including streaming the ~65MB JSON body when the residential proxy fallback is used. |
| CK pricelist fetch timeout | 14m | `ckPricelistFetchTimeout` in `api/gateway/cardkingdom/pricelist_fetch.go` | Context deadline for download + JSON decode + cheapest-listing aggregation. |
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
2. Prefer a single **named constant** in code (e.g. `config.PerSiteTimeout`, `config.DedicatedSearchAttemptTimeout`) and reference that name here. When a store hardcodes a timeout, document the literal and source file.
3. Distinguish **per-request colly policy** (no retry) from **BinderPOS multi-strategy fallback** (up to **four** strategies when a GraphQL token is configured, **two** without GraphQL; one try each per strategy step).
4. When adding a store, update the per-store tables in the BinderPOS or non-BinderPOS section and register it in `api/controller/search.go`.
