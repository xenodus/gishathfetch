/**
 * @typedef {{ store: string, error: string, statusCode?: number }} StoreError
 */

/**
 * @param {StoreError[]} storeErrors
 * @returns {string}
 */
export function formatStoreErrorsSummary(storeErrors) {
  if (!storeErrors?.length) {
    return "";
  }

  const storeNames = storeErrors.map((entry) => entry.store).join(", ");
  const count = storeErrors.length;
  const noun = count === 1 ? "store" : "stores";
  return `${count} ${noun} couldn't be searched: ${storeNames}`;
}

/**
 * @param {StoreError[]} storeErrors
 * @returns {string}
 */
export function formatStoreErrorDetail({ store, error, statusCode }) {
  if (statusCode && !error.includes(`(${statusCode})`)) {
    return `${store}: ${error}`;
  }
  return `${store}: ${error}`;
}
