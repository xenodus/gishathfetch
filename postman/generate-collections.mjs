#!/usr/bin/env node
/**
 * Generates Postman v2.1 collections for each LGS that uses JSON/GraphQL APIs
 * (excludes HTML-only scrapers like Agora Hobby).
 */

import { writeFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';

const OUT_DIR = join(import.meta.dirname, 'collections');
mkdirSync(OUT_DIR, { recursive: true });

const BINDERPOS_GRAPHQL_QUERY = `query SearchCards($q: String!) {
  search(
    query: $q
    first: 25
    types: PRODUCT
    productFilters: [{ available: true }]
  ) {
    edges {
      node {
        ... on Product {
          title
          handle
          availableForSale
          productType
          tags
          featuredImage { url }
          variants(first: 20) {
            edges {
              node {
                id
                title
                availableForSale
                price { amount }
              }
            }
          }
        }
      }
    }
  }
}`;

const FIVEMANA_GRAPHQL_QUERY = `query SearchCards($q: String!) {
  search(
    query: $q
    first: 25
    types: PRODUCT
    productFilters: [{ available: true }, { productType: "MTG Single" }]
  ) {
    edges {
      node {
        ... on Product {
          title
          handle
          availableForSale
          productType
          tags
          featuredImage { url }
          variants(first: 20) {
            edges {
              node {
                title
                availableForSale
                price { amount }
              }
            }
          }
        }
      }
    }
  }
}`;

const CARDS_AND_COLLECTIONS_BODY = {
  query: {
    bool: {
      should: [
        {
          simple_query_string: {
            query: '{{search_query}}',
            fields: ['name', 'setCode', 'setName'],
            default_operator: 'AND',
          },
        },
        {
          multi_match: {
            query: '{{search_query}}',
            type: 'phrase_prefix',
            fields: ['name', 'setCode', 'setName'],
          },
        },
      ],
    },
  },
  post_filter: {
    bool: {
      must: {
        terms: {
          'collectableContext.raw': ['MTG', 'ACCESSORY'],
        },
      },
    },
  },
  aggs: {
    productCategory4: {
      filter: {
        bool: {
          must: {
            terms: {
              'collectableContext.raw': ['MTG', 'ACCESSORY'],
            },
          },
        },
      },
      aggs: {
        'productCategory.raw': {
          terms: { field: 'productCategory.raw', size: 50 },
        },
        'productCategory.raw_count': {
          cardinality: { field: 'productCategory.raw' },
        },
      },
    },
  },
  size: 20,
  sort: [{ quantityOnSale: 'desc' }],
};

function collection(info, variables, items) {
  return {
    info: {
      name: info.name,
      description: info.description,
      schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json',
    },
    variable: variables,
    item: items,
  };
}

function getRequest(name, description, method, url, headers, body) {
  const req = {
    name,
    request: {
      method,
      header: headers,
      url,
      description,
    },
  };
  if (body !== undefined) {
    req.request.body = body;
  }
  return req;
}

function jsonHeaders(extra = []) {
  return [
    { key: 'Accept', value: 'application/json' },
    { key: 'Content-Type', value: 'application/json' },
    ...extra,
  ];
}

function shopifyGraphQLRequest(storeName, baseUrl, token, query) {
  const headers = jsonHeaders([
  { key: 'X-Shopify-Storefront-Access-Token', value: token },
  { key: 'Cookie', value: 'cart_currency=SGD; localization=SG' },
  ]);
  return getRequest(
    'Shopify Storefront GraphQL Search',
    `Search ${storeName} inventory via Shopify Storefront API (same path used by gishathfetch gateway).`,
    'POST',
    {
      raw: '{{base_url}}/api/2024-10/graphql.json',
      host: ['{{base_url}}'],
      path: ['api', '2024-10', 'graphql.json'],
    },
    headers,
    {
      mode: 'raw',
      raw: JSON.stringify({ query, variables: { q: '{{search_query}}' } }, null, 2),
    },
  );
}

function decklistRequest(storeName, shopifyDomain) {
  return getRequest(
    'BinderPOS Decklist API Search',
    `Search ${storeName} via BinderPOS portal decklist API (storeUrl=${shopifyDomain}).`,
    'POST',
    {
      raw: 'https://portal.binderpos.com/external/shopify/decklist?storeUrl={{shopify_domain}}&type=mtg',
      protocol: 'https',
      host: ['portal', 'binderpos', 'com'],
      path: ['external', 'shopify', 'decklist'],
      query: [
        { key: 'storeUrl', value: '{{shopify_domain}}' },
        { key: 'type', value: 'mtg' },
      ],
    },
    jsonHeaders(),
    {
      mode: 'raw',
      raw: JSON.stringify([{ card: '{{search_query}}', quantity: 1 }], null, 2),
    },
  );
}

const binderposStores = [
  { id: 'arcanesanctum', name: 'Arcane Sanctum', baseUrl: 'https://arcanesanctumtcg.com', token: '228ce7e7cffe6623f36634d0ca085e9e', shopifyDomain: '30uetm-1y.myshopify.com' },
  { id: 'cardaffinity', name: 'Card Affinity', baseUrl: 'https://card-affinity.com', token: null, shopifyDomain: '563304-2.myshopify.com' },
  { id: 'cardscitadel', name: 'Cards Citadel', baseUrl: 'https://cardscitadel.com', token: 'b68bd33b7d819fc110eb25a07988cc8e', shopifyDomain: 'card-citadel.myshopify.com' },
  { id: 'flagship', name: 'Flagship Games', baseUrl: 'https://www.flagshipgames.sg', token: 'e08d01a75c052a2ba9e40b5cfa5b0e36', shopifyDomain: 'flagship-games.myshopify.com' },
  { id: 'fyendalhobby', name: 'Fyendal Hobby', baseUrl: 'https://fyendalhobby.com', token: '62ebf8066d0372bca57bb96d7a009a79', shopifyDomain: 'fyendal-hobby.myshopify.com' },
  { id: 'gameshaven', name: 'Games Haven', baseUrl: 'https://www.gameshaventcg.com', token: '5938b052bbbd595d317fdeb5464a6733', shopifyDomain: 'games-haven-sg.myshopify.com' },
  { id: 'gog', name: 'Grey Ogre Games', baseUrl: 'https://www.greyogregames.com', token: '80c454d63abad6ad0ebb6b3aaf649fcd', shopifyDomain: 'grey-ogre-games-singapore.myshopify.com' },
  { id: 'hideout', name: 'Hideout', baseUrl: 'https://hideoutcg.com', token: '986e6f452e7b632be7a14cba965f64a8', shopifyDomain: '220022-20.myshopify.com' },
  { id: 'hideyoshi', name: 'Hideyoshi', baseUrl: 'https://hideyoshitcg.com', token: '08d34daee87c87da96bab125075f193a', shopifyDomain: 'bposacct-9.myshopify.com' },
  { id: 'manapro', name: 'Mana Pro', baseUrl: 'https://sg-manapro.com', token: 'ba695fbdd730f9fa7b1e8f32e36691cf', shopifyDomain: 'mana-pro-sg.myshopify.com' },
  { id: 'mtgasia', name: 'MTG Asia', baseUrl: 'https://www.mtg-asia.com', token: '199d9ab7d26aaba337bd43110a51265f', shopifyDomain: 'mtgasia.myshopify.com' },
  { id: 'onemtg', name: 'OneMtg', baseUrl: 'https://onemtg.com.sg', token: '84d91e8d0ae281e71b204f4c5a8101df', shopifyDomain: 'one-mtg.myshopify.com' },
];

const customStores = [
  {
    id: 'cardscentral',
    name: 'Cards Central',
    description: 'Custom LGS search API at /api/lgs/search. No authentication required.',
    variables: [
      { key: 'base_url', value: 'https://cardscentral.com' },
      { key: 'search_query', value: 'Lightning Bolt' },
    ],
    items: [
      getRequest(
        'LGS Search',
        'GET /api/lgs/search?q={card name}. Returns JSON array of in-stock MTG singles.',
        'GET',
        {
          raw: '{{base_url}}/api/lgs/search?q={{search_query}}',
          host: ['{{base_url}}'],
          path: ['api', 'lgs', 'search'],
          query: [{ key: 'q', value: '{{search_query}}' }],
        },
        [{ key: 'Accept', value: 'application/json' }],
      ),
    ],
  },
  {
    id: 'cardsandcollection',
    name: 'Cards & Collections',
    description: 'Elasticsearch-style catalog API at POST /api/catalog/. No authentication required.',
    variables: [
      { key: 'base_url', value: 'https://cardsandcollections.com' },
      { key: 'search_query', value: 'Counterspell' },
    ],
    items: [
      getRequest(
        'Catalog Search',
        'POST Elasticsearch query to /api/catalog/. Filters to MTG and accessories.',
        'POST',
        {
          raw: '{{base_url}}/api/catalog/',
          host: ['{{base_url}}'],
          path: ['api', 'catalog', ''],
        },
        jsonHeaders(),
        { mode: 'raw', raw: JSON.stringify(CARDS_AND_COLLECTIONS_BODY, null, 2) },
      ),
    ],
  },
  {
    id: 'duellerpoint',
    name: "Dueller's Point",
    description: 'REST search API at GET /products/search. No authentication required.',
    variables: [
      { key: 'base_url', value: 'https://www.duellerspoint.com' },
      { key: 'search_query', value: 'Lightning Bolt' },
    ],
    items: [
      getRequest(
        'Product Search',
        'GET /products/search?search_text={card name}. Returns JSON with results array.',
        'GET',
        {
          raw: '{{base_url}}/products/search?search_text={{search_query}}',
          host: ['{{base_url}}'],
          path: ['products', 'search'],
          query: [{ key: 'search_text', value: '{{search_query}}' }],
        },
        [{ key: 'Accept', value: 'application/json' }],
      ),
    ],
  },
  {
    id: 'moxandlotus',
    name: 'Mox & Lotus',
    description: 'REST product API at GET /api/products. No authentication required.',
    variables: [
      { key: 'base_url', value: 'https://moxandlotus.sg' },
      { key: 'search_query', value: 'Abrade' },
    ],
    items: [
      getRequest(
        'Product Search',
        'GET /api/products with MTG category filters and in_stock=true.',
        'GET',
        {
          raw: '{{base_url}}/api/products?limit=24&full_search=true&showStatus=false&is_paginated=true&in_stock=true&sortVariation=true&category_id=1&variation_code=all&order_by=Price Low to High&search={{search_query}}',
          host: ['{{base_url}}'],
          path: ['api', 'products'],
          query: [
            { key: 'limit', value: '24' },
            { key: 'full_search', value: 'true' },
            { key: 'showStatus', value: 'false' },
            { key: 'is_paginated', value: 'true' },
            { key: 'in_stock', value: 'true' },
            { key: 'sortVariation', value: 'true' },
            { key: 'category_id', value: '1' },
            { key: 'variation_code', value: 'all' },
            { key: 'order_by', value: 'Price Low to High' },
            { key: 'search', value: '{{search_query}}' },
          ],
        },
        [{ key: 'Accept', value: 'application/json' }],
      ),
    ],
  },
  {
    id: 'tcgmarketplace',
    name: 'The TCG Marketplace',
    description: 'Advanced search API on port 3501. Requires access_token (set TCG_MARKETPLACE_ACCESS_TOKEN env var in production).',
    variables: [
      { key: 'api_url', value: 'https://thetcgmarketplace.com:3501' },
      { key: 'access_token', value: '' },
      { key: 'search_query', value: 'Abrade' },
    ],
    items: [
      getRequest(
        'Advanced Search',
        'POST /encoder/advancedsearch. category=3 is MTG. Set access_token collection variable before sending.',
        'POST',
        {
          raw: '{{api_url}}/encoder/advancedsearch',
          host: ['{{api_url}}'],
          path: ['encoder', 'advancedsearch'],
        },
        jsonHeaders(),
        {
          mode: 'raw',
          raw: JSON.stringify(
            {
              access_token: '{{access_token}}',
              name: '{{search_query}}',
              category: 3,
              order: 'name_asc',
            },
            null,
            2,
          ),
        },
      ),
    ],
  },
  {
    id: 'fivemana',
    name: '5 Mana',
    description: 'Shopify Storefront GraphQL (primary path). Filters to available MTG Single products.',
    variables: [
      { key: 'base_url', value: 'https://5-mana.sg' },
      { key: 'storefront_access_token', value: '9e4cb078af6a814458ce898eb9631fe6' },
      { key: 'search_query', value: 'Abrade' },
    ],
    items: [
      getRequest(
        'Shopify Storefront GraphQL Search',
        'POST /api/2024-10/graphql.json with MTG Single product filter.',
        'POST',
        {
          raw: '{{base_url}}/api/2024-10/graphql.json',
          host: ['{{base_url}}'],
          path: ['api', '2024-10', 'graphql.json'],
        },
        jsonHeaders([
          { key: 'X-Shopify-Storefront-Access-Token', value: '{{storefront_access_token}}' },
          { key: 'Cookie', value: 'cart_currency=SGD; localization=SG' },
        ]),
        {
          mode: 'raw',
          raw: JSON.stringify({ query: FIVEMANA_GRAPHQL_QUERY, variables: { q: '{{search_query}}' } }, null, 2),
        },
      ),
    ],
  },
];

function writeCollection(filename, data) {
  const path = join(OUT_DIR, filename);
  writeFileSync(path, JSON.stringify(data, null, 2) + '\n');
  return path;
}

const written = [];

for (const store of customStores) {
  const c = collection(
    {
      name: `${store.name} - LGS Search API`,
      description: store.description + '\n\nGenerated from gishathfetch gateway definitions. API paths only (no HTML scrape fallbacks).',
    },
    store.variables,
    store.items,
  );
  written.push(writeCollection(`${store.id}.postman_collection.json`, c));
}

for (const store of binderposStores) {
  const variables = [
    { key: 'base_url', value: store.baseUrl },
    { key: 'shopify_domain', value: store.shopifyDomain },
    { key: 'search_query', value: 'Opt' },
  ];
  if (store.token) {
    variables.push({ key: 'storefront_access_token', value: store.token });
  }

  const items = [];
  if (store.token) {
    items.push(
      shopifyGraphQLRequest(store.name, store.baseUrl, '{{storefront_access_token}}', BINDERPOS_GRAPHQL_QUERY),
    );
  }
  items.push(decklistRequest(store.name, store.shopifyDomain));

  const apiNote = store.token
    ? 'Includes Shopify Storefront GraphQL and BinderPOS Decklist API requests.'
    : 'Card Affinity has no public Storefront token; only the BinderPOS Decklist API is available.';

  const c = collection(
    {
      name: `${store.name} - LGS Search API`,
      description: `${apiNote}\n\nGenerated from gishathfetch gateway definitions. API paths only (no HTML scrape fallbacks).`,
    },
    variables,
    items,
  );
  written.push(writeCollection(`${store.id}.postman_collection.json`, c));
}

// Master collection with folders per store
const masterItems = [
  ...customStores.map((s) => ({
    name: s.name,
    item: s.items,
    description: s.description,
  })),
  ...binderposStores.map((s) => {
    const subItems = [];
    if (s.token) {
      subItems.push(shopifyGraphQLRequest(s.name, s.baseUrl, s.token, BINDERPOS_GRAPHQL_QUERY));
    }
    subItems.push(decklistRequest(s.name, s.shopifyDomain));
    return { name: s.name, item: subItems };
  }),
];

const master = collection(
  {
    name: 'GishathFetch - All LGS Search APIs',
    description:
      'Combined Postman collection for all API-based LGS integrations in gishathfetch. Excludes HTML-only stores (Agora Hobby). Each folder mirrors an individual per-store collection in postman/collections/.',
  },
  [{ key: 'search_query', value: 'Opt' }],
  masterItems,
);

written.push(writeCollection('all-lgs.postman_collection.json', master));

console.log(`Wrote ${written.length} Postman collections to ${OUT_DIR}:`);
for (const p of written) {
  console.log('  -', p.split('/').pop());
}
