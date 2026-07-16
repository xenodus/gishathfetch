# BinderPOS search feature parity

This document records **feature parity** between the three BinderPOS data sources (HTML scrape, Shopify Storefront GraphQL, BinderPOS Decklist API), the four HTML layout variants, and the transport modes that compose each named strategy step.

For **strategy order, timeouts, proxy config, and fallback rules**, see [`search-strategies-retries-timeouts.md`](search-strategies-retries-timeouts.md). Update this file when changing card-field mapping, MTG filtering, result limits, or scrap-variant behavior.

---

## Overview

All 11 BinderPOS-backed stores share one orchestrator: `binderpos.impl.Search` in `api/gateway/binderpos/storefront_search.go`. It runs an **ordered list of strategy steps** (not retries of a single request). Each step is one HTTP interaction bounded by `binderposAttemptTimeout` (5s).

| Data source | Implementation | Named strategies |
|-------------|----------------|------------------|
| **Shopify Storefront GraphQL** | `storefront_graphql.go` | `graphql-dedicated`, `graphql-direct` |
| **HTML scrape** | `scrap.go`, `scrap_dynamic.go` | `scrap-dedicated`, `scrap-direct`, `scrap-dynamic` |
| **BinderPOS Decklist API** | `storefront_decklist.go`, `storefront_client.go` | `decklist-dedicated`, `decklist-direct`, `decklist-dynamic` |

**Transport modes** (dedicated proxy, direct, dynamic proxy) are orthogonal to the data source. GraphQL has **no dynamic-proxy step**; dynamic proxy is reserved for the final scrap and decklist fallbacks.

---

## Strategy order and selection

### With Storefront access token (8 steps)

`graphql-dedicated` → `graphql-direct` → `scrap-dedicated` → `scrap-direct` → `decklist-dedicated` → `decklist-direct` → `scrap-dynamic` → `decklist-dynamic`

### Without Storefront access token (6 steps)

`scrap-dedicated` → `scrap-direct` → `decklist-dedicated` → `decklist-direct` → `scrap-dynamic` → `decklist-dynamic`

Only **Card Affinity** omits GraphQL (no token configured).

### Fallback rules (`storefront_fallback.go`)

| Outcome | Next step |
|---------|-----------|
| Returns **cards** | Stop — success |
| **Scrap or GraphQL**: empty, no error | **Final** — return empty (decklist not tried) |
| **Decklist**: empty, no error | Skip remaining **decklist** steps only |
| **Scrap or GraphQL**: HTTP **5xx** | **Final** — return error |
| **Scrap or GraphQL**: other error (403, 429, network) | Try next strategy |
| **Decklist**: error | Try next strategy (unless decklist family abandoned) |
| All attempts fail | Return last annotated error: `attempt N (strategy-name): …` |

The first strategy skips per-domain pacing; later steps honor the 200ms minimum interval per host.

---

## Stores and HTML scrap variants

Each store package (`api/gateway/{store}/search.go`) passes a **scrap variant** (1–4) that selects the HTML parser and influences card-name formatting for API methods too.

| Store | Scrap variant | GraphQL token | Shopify domain |
|-------|---------------|---------------|----------------|
| Card Affinity | 2 | No | `563304-2.myshopify.com` |
| Cards Citadel | 1 | Yes | `card-citadel.myshopify.com` |
| Flagship | 2 | Yes | `flagship-games.myshopify.com` |
| Fyendal Hobby | 4 | Yes | `fyendal-hobby.myshopify.com` |
| Games Haven | 3 | Yes | `games-haven-sg.myshopify.com` |
| GOG | 3 | Yes | `grey-ogre-games-singapore.myshopify.com` |
| Hideout | 3 | Yes | `220022-20.myshopify.com` |
| Hideyoshi | 2 | Yes | `bposacct-9.myshopify.com` |
| Mana Pro | 2 | Yes | `mana-pro-sg.myshopify.com` |
| MTG Asia | 2 | Yes | `mtgasia.myshopify.com` |
| One MTG | 2 | Yes | `one-mtg.myshopify.com` |

### HTML variant details

| Variant | Stores | Search query shaping | DOM / data source | MTG filter on scrape path |
|---------|--------|----------------------|-------------------|---------------------------|
| **1** | Cards Citadel | `searchStr + " mtg"`; wildcard template `/search?q=*%s*` | `div.Norm` + `div.addNow` price text | Query suffix only |
| **2** | Card Affinity, Flagship, Hideyoshi, Mana Pro, MTG Asia, One MTG | `searchStr + " mtg"` | Embedded `data-product-variants` JSON (`CardInfo`) | `shouldIncludeBinderposProduct` when `data-product-type` present |
| **3** | Games Haven, GOG, Hideout | `searchStr + " mtg"` | `div.productCard__card` + `ul.productChip__grid` variant chips | Query suffix only |
| **4** | Fyendal Hobby | `product_type:"MTG Single Cards" AND {searchStr}` | `div.product-item--vertical` theme layout | Server-side product_type filter in query |

---

## Data source endpoints

| Source | Endpoint | Auth / params |
|--------|----------|---------------|
| **GraphQL** | `{storeBaseURL}/api/2024-10/graphql.json` | Header `X-Shopify-Storefront-Access-Token`; body is Storefront `search` query |
| **Decklist** | `POST https://portal.binderpos.com/external/shopify/decklist?storeUrl={shopifyDomain}&type=mtg` | JSON body `[{"card": searchStr, "quantity": 1}]`; portal concurrency capped at 4 in-flight |
| **HTML scrape** | Store-specific `searchURL` on the public storefront | Colly single-page fetch; SGD currency cookie applied |

`doDecklistRequestWithRetry` (`storefront_decklist_retry.go`) sends **one** HTTP request despite the name — no automatic retry on 429/5xx.

---

## Returned card model (`gateway.Card`)

All methods map into the same struct (`api/gateway/spec.go`):

| Field | Description |
|-------|-------------|
| `Name` | Display name (formatting depends on scrap variant — see below) |
| `Url` | Product URL with `?variant={id}&utm_source=` |
| `Img` | Product image or placeholder |
| `Price` | SGD float |
| `InStock` | Availability flag |
| `IsFoil` | Foil finish |
| `Source` | Store display name |
| `Quality` | Condition (NM/LP/etc. via `util.MapQuality` where applied) |
| `ExtraInfo` | Optional; set name for scrap variant 3 stores |

**Never returned by any BinderPOS method:** language, quantity (decklist reads `quantity` but only uses it for stock gating), set as a dedicated field (may appear in `Name` or `ExtraInfo`).

---

## Field population parity

### Cross-method summary

| Capability | GraphQL | Decklist | HTML scrape |
|------------|:-------:|:--------:|:-----------:|
| Requires Storefront token | Yes | No | No |
| Requires Shopify domain | No | Yes | No |
| MTG scope | Client: `isMagicProductType` | Server: `type=mtg` | Query suffix / product_type / variant-2 DOM filter |
| Availability filter | Server: `available: true` | `quantity > 0` | Per-variant DOM checks |
| Max products | **25** (`first: 25`) | Single-card lookup | Single HTML page |
| Max variants per product | **20** (`variants(first: 20)`) | All in API response | All in DOM/JSON |
| Pagination | No | No | No (unused `pagination` struct in `scrap.go`) |
| Client-side dedup | No | Yes (`url\|price\|instock`) | No |
| `MapQuality` on condition | Yes | Yes | Variant-dependent (see below) |
| `ExtraInfo` (set) | Variant 3 only | Variant 3 only | Variant 3 HTML only |
| SGD currency cookie | Yes | No | Yes (colly `OnRequest`) |
| Dynamic proxy transport | **No** | Yes | Yes |
| Portal concurrency gate | N/A | 4 in-flight | N/A |

### Per-field mapping

| Field | GraphQL | Decklist | Scrap v1 | Scrap v2 | Scrap v3 | Scrap v4 |
|-------|---------|----------|----------|----------|----------|----------|
| **Name** | `formatCardName` | `formatCardName` | `p.productTitle` | `card.Name` (JSON) | `p.productCard__title` | `parseFyendalNameAndFoil` |
| **Url** | `buildProductURLWithVariant` | same | same | same | same | same |
| **Img** | Featured image or placeholder | API `img` or placeholder | `https:` + `img[src]` | `https:` + src (placeholder if no-image) | `https:` + `data-src` | `img.product-item__primary-image` |
| **Price** | `variant.price.amount` | `variant.price` | Parsed from add-now text | `card.Price / 100` | `data-variantprice / 100` | `util.ParsePrice` |
| **InStock** | `availableForSale` (always true when returned) | `quantity > 0` | add-now non-empty | `card.Available` | variant available + qty ≠ 0 | not "Sold out" |
| **IsFoil** | variant title + `titleIndicatesFoil` | variant title contains "foil" | quality string | variant title | `data-varianttitle` | `[foil]` title prefix |
| **Quality** | `MapQuality(variant.Title)` | `MapQuality(variant.Title)` | parsed from price string | raw title minus "Foil" (**no MapQuality**) | `MapQuality(chip text)` | **not set** |
| **ExtraInfo** | `extractSetName` if variant 3 | `setName` or extracted if variant 3 | — | — | `p.productCard__setName` | — |

### Name formatting by scrap variant (`formatCardName`)

Shared by GraphQL and Decklist; HTML scrap uses variant-specific parsers instead.

| Variant | Rule |
|---------|------|
| **2** | `"Product Title - Variant Title"` |
| **3** | Strip trailing `[Set Name]` from product title |
| **1, 4** | Product title as-is |

---

## Search query shaping

| Method | Query passed to upstream |
|--------|--------------------------|
| **GraphQL** | Raw `searchStr` (MTG filtered client-side after fetch) |
| **Decklist** | Raw `searchStr` in JSON `card` field (`type=mtg` scopes server-side) |
| **HTML v1–v3** | `searchStr + " mtg"` appended before URL encoding |
| **HTML v4 (Fyendal)** | `product_type:"MTG Single Cards" AND {searchStr}` |

Cards Citadel (variant 1) additionally uses a wildcard search template: `/search?q=*%s*`.

---

## MTG product filtering

| Path | Mechanism | Source |
|------|-----------|--------|
| GraphQL | `isMagicProductType(productType, tags)` — drops non-MTG products after fetch | `product_filter.go` |
| Decklist | `type=mtg` query parameter on portal API | `storefront.go` |
| Scrap variant 2 | `shouldIncludeBinderposProduct` on `data-product-type` / `data-product-tags` when attribute present | `product_filter.go` |
| Scrap v1, v3 | `" mtg"` query suffix | `scrap.go` |
| Scrap v4 | `product_type:"MTG Single Cards"` in search query | `scrap.go` |

`isMagicProductType` matches product types containing `mtg` or `magic the gathering`, or tags equal to `MTG`.

---

## Known gaps and intentional differences

1. **GraphQL has no dynamic-proxy path** — only dedicated and direct transports.
2. **Result breadth differs** — GraphQL caps at 25 products × 20 variants; decklist is a single-card portal lookup; scrape is one HTML results page with no pagination.
3. **Empty scrap/GraphQL is final** — a successful empty HTML or GraphQL response prevents decklist from running, even though decklist might have found stock.
4. **Quality normalization inconsistency** — scrap variant 2 skips `MapQuality`; variant 4 omits `Quality` entirely.
5. **Set metadata** — only scrap variant 3 (and API methods when `scrapVariant == 3`) populate `ExtraInfo` with set name.
6. **Decklist deduplication only** — other methods may return duplicate `url|price|instock` combinations.
7. **Quantity not exposed** — decklist has per-variant stock counts but they are not mapped to `gateway.Card`.
8. **Language not supported** — no method extracts or returns card language.

---

## Tests

| File | What it covers |
|------|----------------|
| `storefront_search_test.go` | Strategy order |
| `storefront_fallback_test.go` | Fallback rules (empty scrap, 5xx, decklist skip, 429 → decklist) |
| `storefront_graphql_test.go` | Variant ID parsing, `mapGraphQLProduct`, GraphQL fallback |
| `storefront_decklist_test.go` | JSON decode, `mapDecklistLinesToCards` |
| `storefront_decklist_retry_test.go` | Single-send decklist HTTP |
| `storefront_proxy_client_test.go` | Proxy round-robin, dynamic proxy env, timeout/pacing |
| `storefront_portal_gate_test.go` | Portal concurrency gate |
| `product_filter_test.go` | MTG product-type filtering |
| `scrap_variant4_test.go` | Fyendal query/name/price helpers |
| `scrap_test.go` | Live scrape per variant (`RUN_BINDERPOS_LIVE_TESTS=1`) |
| Per-store `search_test.go` | Live search or HTML structure probe via `gatewaytest.RequireSearchOrProbe` |

---

## Key source files

| Area | Path |
|------|------|
| Public API | `api/gateway/binderpos/new.go` |
| Orchestration | `api/gateway/binderpos/storefront_search.go` |
| Fallback logic | `api/gateway/binderpos/storefront_fallback.go` |
| HTML scrape | `api/gateway/binderpos/scrap.go`, `scrap_dynamic.go` |
| GraphQL | `api/gateway/binderpos/storefront_graphql.go` |
| Decklist | `api/gateway/binderpos/storefront_decklist.go`, `storefront_decklist_retry.go` |
| HTTP clients | `api/gateway/binderpos/storefront_client.go` |
| Card mapping | `api/gateway/binderpos/storefront_card_helpers.go` |
| MTG filter | `api/gateway/binderpos/product_filter.go` |
| Store registry | `api/controller/search.go` |
