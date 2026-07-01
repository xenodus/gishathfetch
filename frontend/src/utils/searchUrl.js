import {
  LGS_OPTIONS,
  POPULAR_SEARCH_UTM_CAMPAIGN,
  POPULAR_SEARCH_UTM_MEDIUM,
  POPULAR_SEARCH_UTM_SOURCE,
} from "../constants";

/**
 * @typedef {{
 *   query: string,
 *   stores: string[],
 *   results: object[],
 *   storeErrors: object[],
 *   hasSearched: boolean,
 *   searchError: string | null,
 *   cardKingdomPrice: object | null,
 * }} SearchHistoryState
 */

/**
 * @param {SearchHistoryState} snapshot
 * @returns {SearchHistoryState}
 */
export function buildSearchHistoryState(snapshot) {
  return {
    query: snapshot.query,
    stores: snapshot.stores,
    results: snapshot.results,
    storeErrors: snapshot.storeErrors,
    hasSearched: snapshot.hasSearched,
    searchError: snapshot.searchError,
    cardKingdomPrice: snapshot.cardKingdomPrice ?? null,
  };
}

/**
 * @param {unknown} state
 * @returns {state is SearchHistoryState}
 */
export function isSearchHistoryState(state) {
  return (
    !!state &&
    typeof state === "object" &&
    "query" in state &&
    "stores" in state &&
    "results" in state
  );
}

/**
 * @param {string} lgsParam
 * @returns {string[]}
 */
export function parseStoresFromUrlParam(lgsParam) {
  if (!lgsParam) {
    return [];
  }

  const stores = lgsParam
    .split(",")
    .map((store) => store.trim())
    .filter((store) => store.length > 0 && LGS_OPTIONS.includes(store));

  return [...new Set(stores)];
}

/**
 * @param {string} key
 * @returns {boolean}
 */
export function isTrackingParam(key) {
  return key.toLowerCase().startsWith("utm_");
}

/**
 * Copy utm_* params from source into target without overwriting keys already set.
 *
 * @param {URLSearchParams} sourceParams
 * @param {URLSearchParams} targetParams
 */
export function mergeTrackingParams(sourceParams, targetParams) {
  for (const [key, value] of sourceParams.entries()) {
    if (isTrackingParam(key) && !targetParams.has(key)) {
      targetParams.set(key, value);
    }
  }
}

/**
 * @param {string} query
 * @param {string[]} stores
 * @param {URLSearchParams} [existingParams]
 * @returns {URLSearchParams}
 */
export function buildSearchUrlParams(query, stores, existingParams) {
  const params = new URLSearchParams();
  params.set("s", query);

  const validStores = stores.filter((store) => LGS_OPTIONS.includes(store));
  const isAllStores =
    validStores.length === LGS_OPTIONS.length &&
    LGS_OPTIONS.every((store) => validStores.includes(store));

  if (validStores.length > 0 && !isAllStores) {
    params.set("lgs", validStores.join(","));
  }

  if (existingParams) {
    mergeTrackingParams(existingParams, params);
  }

  return params;
}

/**
 * @param {string} baseUrl
 * @param {string} query
 * @param {string[]} stores
 * @param {URLSearchParams} [existingParams]
 * @returns {string}
 */
export function buildSearchUrl(baseUrl, query, stores, existingParams) {
  const params = buildSearchUrlParams(query, stores, existingParams);
  return `${baseUrl}?${params.toString()}`;
}

/**
 * @param {string} baseUrl
 * @param {string} query
 * @returns {string}
 */
export function buildSearchQueryUrl(baseUrl, query) {
  const params = new URLSearchParams();
  params.set("s", query);
  return `${baseUrl}?${params.toString()}`;
}

/**
 * @param {string} baseUrl
 * @param {string} query
 * @param {string} [period]
 * @returns {string}
 */
export function buildPopularSearchUrl(baseUrl, query, period) {
  const params = new URLSearchParams();
  params.set("s", query);
  params.set("utm_source", POPULAR_SEARCH_UTM_SOURCE);
  params.set("utm_medium", POPULAR_SEARCH_UTM_MEDIUM);
  params.set("utm_campaign", POPULAR_SEARCH_UTM_CAMPAIGN);
  if (period) {
    params.set("utm_content", period);
  }
  return `${baseUrl}?${params.toString()}`;
}

/**
 * @param {URLSearchParams} urlParams
 * @returns {string[] | null}
 */
export function getStoresFromUrl(urlParams) {
  if (
    urlParams.has("src") &&
    LGS_OPTIONS.includes(decodeURIComponent(urlParams.get("src")))
  ) {
    return [decodeURIComponent(urlParams.get("src"))];
  }

  if (urlParams.has("lgs")) {
    const stores = parseStoresFromUrlParam(
      decodeURIComponent(urlParams.get("lgs")),
    );
    return stores.length > 0 ? stores : null;
  }

  return null;
}

/**
 * @returns {boolean}
 */
export function hasStoredStoreSelection() {
  try {
    return localStorage.getItem("lgsSelected") !== null;
  } catch {
    return false;
  }
}

/**
 * @param {string[]} stores
 */
export function persistSelectedStores(stores) {
  try {
    localStorage.setItem("lgsSelected", encodeURIComponent(stores.join(",")));
  } catch (err) {
    console.error("Failed to save selected stores:", err);
  }
}

/**
 * @param {URLSearchParams} [urlParams]
 * @returns {string[]}
 */
export function getInitialSelectedStores(
  urlParams = new URLSearchParams(window.location.search),
) {
  const urlStores = getStoresFromUrl(urlParams);
  if (urlStores) {
    persistSelectedStores(urlStores);
    return urlStores;
  }

  const storedLgs = localStorage.getItem("lgsSelected");
  if (storedLgs !== null) {
    const decoded = decodeURIComponent(storedLgs);
    return decoded === "" ? [] : decoded.split(",");
  }

  return LGS_OPTIONS;
}
