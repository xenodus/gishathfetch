# Gishath Fetch

Gishath Fetch is a web application for Magic: The Gathering players in Singapore to search singles across multiple local game stores (LGS) in parallel.

It aggregates listings from supported stores, normalizes results, and sorts by price so users can quickly find the best available options.

## 🚀 Features

- ⚡ Concurrent search across supported stores
- 🎯 Result filtering and normalization for better match quality
- 💰 Price-first sorting for faster deal discovery
- 🧭 Store filtering (query specific LGS only)
- 🛒 Persistent cart in the frontend UI, sharable across devices via export/import code
- 📈 Trending searches — popular card names by time range (24 hours to 1 year), plus top Card Kingdom price risers and drops (24h)
- 🏷️ Card Kingdom price reference on search results (USD retail from a daily-updated index)

## 🏗️ Architecture

Gishath Fetch has three main layers plus external integrations. Full diagrams,
service tables, batch-job flows, and IAM notes:
[`docs/architecture.md`](docs/architecture.md).

| Layer | Location | Summary |
|-------|----------|---------|
| **Frontend** | `frontend/` | React 19 + Vite + Bootstrap SPA — search UI, cart, trending keywords |
| **Backend** | `api/` | Go Lambda handlers, search controller, per-store scraper gateways |
| **Infrastructure** | AWS ap-southeast-1 | CloudFront, WAF, S3, API Gateway, Lambda, DynamoDB, EventBridge, ECR |

The browser loads the SPA from **S3 via CloudFront** (`gishathfetch.com`). Search
and session call **`api.gishathfetch.com`** (separate CloudFront → API Gateway →
Lambda path). Daily EventBridge jobs refresh Card Kingdom prices in DynamoDB and
export GA4 trending keywords to S3. Both CloudFront distributions sit behind **AWS
WAF**; inbound API abuse mitigation is documented in
[`docs/api-abuse-mitigation.md`](docs/api-abuse-mitigation.md).

## 🛡️ API abuse mitigation

Inbound `/search` and `/session` are gated by two optional layers (each off
until configured), with **AWS WAF** on both CloudFront distributions at the edge:

1. **CloudFront → API origin secret** — Lambda `API_ORIGIN_VERIFY_SECRET`;
   CloudFront (or the Vite dev proxy) injects `X-Origin-Verify` on origin
   requests.
2. **Session token** — `GET /session` mints HttpOnly `gf_api_session`
   (HMAC via `API_SESSION_SECRET`, default TTL 15m); `/search` requires it.

Details, env reference, CloudFront header setup, and the browser sequence
diagram: [`docs/api-abuse-mitigation.md`](docs/api-abuse-mitigation.md).

## 🔎 Search flow

A search request fans out to every selected store in parallel, each store
resolves its own listings, and the results are merged, filtered, and sorted
before being returned. Callers must pass abuse-mitigation checks first (see
above) when those controls are enabled in the environment.

### Request entry & fan-out

1. The handler enforces origin verify and (for `/search`) the session cookie
   when configured, then parses `s` (the search string, minimum 3 characters) and an
   optional `lgs` filter (comma-separated store names; empty means all stores).
2. The controller instantiates each selected store and runs **one goroutine per
   store**, each bounded by a 20s per-site timeout (`config.PerSiteTimeout`).
3. Each store's results are merged into a shared aggregator. A per-store failure
   is recorded but never blocks the others, so a search returns whatever
   succeeded (partial success).
4. The aggregated cards are filtered and sorted: **in-stock only**, **price
   ascending**, with name-match priority **exact > prefix > partial**. Art cards
   and Japanese-language listings are excluded. A minimum response time (~1s) is
   enforced for a consistent UX.

### Two kinds of stores

**Non-BinderPOS stores** (e.g. Agora Hobby, Cards Central, Cards & Collections,
Dueller's Point, Mox & Lotus, The TCG Marketplace) each implement a single
bespoke `Search` — a custom JSON API call or one HTML scrape — with no
multi-strategy fallback chain (5 Mana is an exception; see below). On failure the
store simply contributes nothing.

**5 Mana** is Shopify (Dawn, not BinderPOS). It tries Storefront GraphQL first,
then falls back to a `main-search` HTML section scrape when GraphQL fails.

**BinderPOS stores** (e.g. Arcane Sanctum, Card Affinity, Cards Citadel, Flagship, Game's Haven,
Fyendal Hobby, Grey Ogre Games, Hideout, Hideyoshi, Mana Pro, MTG Asia,
OneMTG) share one gateway. Stores with a configured Storefront access token try
**GraphQL first** (dedicated → direct), then fall back to HTML scrape.

### BinderPOS GraphQL and scrape fallback chain

When a store has a Storefront access token:

`graphql-dedicated` → `graphql-direct` → `scrap-dedicated` → `scrap-direct`

Without a token, the chain starts at `scrap-dedicated`. GraphQL uses dedicated then direct only.

The BinderPOS Decklist API remains implemented under `api/gateway/binderpos/` for
reference and Postman collections, but it is not part of the live search fallback
chain.

```mermaid
flowchart TD
    A[BinderPOS store search] --> G1[graphql-dedicated]
    G1 -- error --> G2[graphql-direct]
    G2 -- error --> B[scrap-dedicated]
    B -- error --> C[scrap-direct]
    G1 -- cards or empty success --> E[Return result]
    G2 -- cards or empty success --> E
    B -- cards or empty success --> E
    C -- cards, empty success, or error --> E
```

### Fallback rules

- The chain advances to the next attempt **on error only**. An empty but
  error-free **GraphQL** or **scrape** result counts as success and stops the
  chain.
- HTTP **5xx** errors on scrape or GraphQL attempts are final. Other GraphQL
  failures fall through to HTML scrap.
- Each attempt is bounded by a 5s timeout (`binderposAttemptTimeout`). The first
  attempt starts immediately; later attempts honor per-domain request pacing.

## 🗂️ Repository layout

```text
.
|-- api/         # Go backend (Lambda handler, scraping gateways, tests)
|-- docs/        # Maintainer docs (architecture, API abuse mitigation, search strategies, skills)
|-- frontend/    # React + Vite single-page app
|-- Makefile     # Local helpers for common project tasks
`-- Dockerfile   # Backend container build definition
```

## ✅ Prerequisites

- Node.js 22 (matches CI workflow)
- npm
- Go (version declared in `api/go.mod`)

## 🧪 Tests

From repo root:

```bash
make test
```

Or directly:

```bash
cd api
go clean -testcache
go test -mod=vendor -failfast -timeout 5m ./...
```

## 🌐 Proxy support (rate limiting)

The scraper supports multiple proxies to reduce rate-limiting issues from upstream stores.

## 📜 License

This project is licensed under the MIT License. See [LICENSE](./LICENSE).

---

Gishath Fetch is not affiliated with Wizards of the Coast or any supported local game store.
