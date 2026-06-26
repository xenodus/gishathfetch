const DB_NAME = "gishathfetch-market";
const DB_VERSION = 1;
const STORE = "blobs";

const openDb = () =>
  new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);
    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE)) {
        db.createObjectStore(STORE);
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });

const withStore = async (mode, fn) => {
  const db = await openDb();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE, mode);
    const store = tx.objectStore(STORE);
    const result = fn(store);
    tx.oncomplete = () => resolve(result);
    tx.onerror = () => reject(tx.error);
  });
};

export const getCachedBlob = async (key) => {
  return withStore("readonly", (store) => {
    return new Promise((resolve, reject) => {
      const request = store.get(key);
      request.onsuccess = () => resolve(request.result ?? null);
      request.onerror = () => reject(request.error);
    });
  });
};

export const setCachedBlob = async (key, value) => {
  return withStore("readwrite", (store) => {
    store.put(value, key);
  });
};

export const getCachedJson = async (key) => {
  const entry = await getCachedBlob(key);
  if (!entry?.data) {
    return null;
  }
  return entry;
};

export const setCachedJson = async (key, data, ttlMs) => {
  await setCachedBlob(key, {
    data,
    expiresAt: Date.now() + ttlMs,
  });
};

export const isCacheFresh = (entry) =>
  Boolean(entry?.data && entry.expiresAt > Date.now());
