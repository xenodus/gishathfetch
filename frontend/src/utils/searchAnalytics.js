/** GA4 event for exact card searches with zero in-stock hits at a store. */
export const LGS_CARD_NOT_FOUND_EVENT = "lgs_card_not_found";

export function foldCardNameForMatch(name) {
  return String(name ?? "")
    .trim()
    .normalize("NFKD")
    .replace(/\p{M}/gu, "")
    .toLowerCase();
}

export function cardNamesMatchForVerify(canonical, query) {
  const left = String(canonical ?? "");
  const right = String(query ?? "").trim();
  if (!left || !right) {
    return false;
  }
  if (left.toLowerCase() === right.toLowerCase()) {
    return true;
  }
  return foldCardNameForMatch(left) === foldCardNameForMatch(right);
}

/**
 * Returns the canonical Scryfall card name when query is a 1:1 match (no fuzzy).
 */
export async function verifyExactCardName(query, signal) {
  const trimmed = String(query ?? "").trim();
  if (!trimmed) {
    return null;
  }

  const autocompleteURL = `https://api.scryfall.com/cards/autocomplete?q=${encodeURIComponent(trimmed)}`;
  const autocompleteRes = await fetch(autocompleteURL, { signal });
  if (autocompleteRes.ok) {
    const autocomplete = await autocompleteRes.json();
    for (const name of autocomplete?.data ?? []) {
      if (cardNamesMatchForVerify(name, trimmed)) {
        return name;
      }
    }
  }

  const namedURL = `https://api.scryfall.com/cards/named?exact=${encodeURIComponent(trimmed)}`;
  const namedRes = await fetch(namedURL, { signal });
  if (namedRes.ok) {
    const card = await namedRes.json();
    if (card?.name && cardNamesMatchForVerify(card.name, trimmed)) {
      return card.name;
    }
  }

  return null;
}

export function listStoresWithNoResults(stores, storeStats, storeErrors) {
  const searchedStores = Array.isArray(stores) ? stores : [];
  if (searchedStores.length === 0) {
    return [];
  }

  const erroredStores = new Set(
    (Array.isArray(storeErrors) ? storeErrors : [])
      .map((entry) => entry?.store)
      .filter(Boolean),
  );
  const itemCountByStore = new Map(
    (Array.isArray(storeStats) ? storeStats : []).map((stat) => [
      stat?.store,
      stat?.itemCount,
    ]),
  );

  return searchedStores.filter((store) => {
    if (!store || erroredStores.has(store)) {
      return false;
    }
    if (!itemCountByStore.has(store)) {
      return false;
    }
    return itemCountByStore.get(store) === 0;
  });
}

function sendGtagEvent(name, params) {
  if (typeof window === "undefined" || !window.gtag) {
    return;
  }
  window.gtag("event", name, params);
}

/**
 * Records GA4 search telemetry for exact card-name queries and tags stores with
 * zero in-stock results so LGS demand can be measured in analytics.
 */
export async function recordSearchAnalytics({
  query,
  stores,
  storeStats,
  storeErrors,
  signal,
}) {
  let canonicalName;
  try {
    canonicalName = await verifyExactCardName(query, signal);
  } catch (err) {
    if (err?.name === "AbortError") {
      return;
    }
    console.warn("Search analytics skipped: Scryfall verify failed", err);
    return;
  }

  if (!canonicalName) {
    return;
  }

  sendGtagEvent("search", { search_term: canonicalName });
  sendGtagEvent("view_search_results", { search_term: canonicalName });

  const storesWithoutStock = listStoresWithNoResults(
    stores,
    storeStats,
    storeErrors,
  );
  for (const store of storesWithoutStock) {
    sendGtagEvent(LGS_CARD_NOT_FOUND_EVENT, {
      search_term: canonicalName,
      lgs: store,
    });
  }
}
