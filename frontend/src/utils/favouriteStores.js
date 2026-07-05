import { LGS_OPTIONS } from "../constants";

export const FAVOURITE_STORES_STORAGE_KEY = "lgsFavourites";

/**
 * @param {string[]} stores
 * @returns {string[]}
 */
export function validateFavouriteStores(stores) {
  if (!Array.isArray(stores)) {
    return [];
  }

  const validStores = new Set(LGS_OPTIONS);
  return stores.filter((store) => validStores.has(store));
}

/**
 * @param {string[]} left
 * @param {string[]} right
 * @returns {boolean}
 */
export function storeSelectionsMatch(left, right) {
  if (left.length !== right.length) {
    return false;
  }

  const leftSet = new Set(left);
  return right.every((store) => leftSet.has(store));
}

/**
 * @returns {string[]}
 */
export function loadFavouriteStoresFromStorage() {
  try {
    const stored = localStorage.getItem(FAVOURITE_STORES_STORAGE_KEY);
    if (stored === null) {
      return [];
    }

    const decoded = decodeURIComponent(stored);
    if (decoded === "") {
      return [];
    }

    return validateFavouriteStores(decoded.split(","));
  } catch (err) {
    console.error("Failed to load favourite stores:", err);
    return [];
  }
}

/**
 * @param {string[]} stores
 */
export function persistFavouriteStores(stores) {
  try {
    const validated = validateFavouriteStores(stores);
    localStorage.setItem(
      FAVOURITE_STORES_STORAGE_KEY,
      encodeURIComponent(validated.join(",")),
    );
  } catch (err) {
    console.error("Failed to save favourite stores:", err);
  }
}

/**
 * @returns {boolean}
 */
export function hasStoredFavouriteStores() {
  try {
    return localStorage.getItem(FAVOURITE_STORES_STORAGE_KEY) !== null;
  } catch {
    return false;
  }
}
