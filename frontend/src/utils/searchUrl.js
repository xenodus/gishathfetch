import { LGS_OPTIONS } from "../constants";

/**
 * @typedef {{
 *   query: string,
 *   stores: string[],
 *   results: object[],
 *   storeErrors: object[],
 *   hasSearched: boolean,
 *   searchError: string | null,
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
 * @param {string} query
 * @param {string[]} stores
 * @returns {URLSearchParams}
 */
export function buildSearchUrlParams(query, stores) {
  const params = new URLSearchParams();
  params.set("s", query);

  const validStores = stores.filter((store) => LGS_OPTIONS.includes(store));
  const isAllStores =
    validStores.length === LGS_OPTIONS.length &&
    LGS_OPTIONS.every((store) => validStores.includes(store));

  if (validStores.length > 0 && !isAllStores) {
    params.set("lgs", validStores.join(","));
  }

  return params;
}

/**
 * @param {string} baseUrl
 * @param {string} query
 * @param {string[]} stores
 * @returns {string}
 */
export function buildSearchUrl(baseUrl, query, stores) {
  const params = buildSearchUrlParams(query, stores);
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
