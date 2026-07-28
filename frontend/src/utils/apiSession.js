import { API_SESSION_URL } from "../constants";

let sessionBootstrapPromise = null;

/**
 * Ensures the browser holds a valid HttpOnly API session cookie (minted by GET /api/session).
 * Safe to call repeatedly; concurrent callers share one in-flight request.
 */
export function ensureApiSession() {
  if (!sessionBootstrapPromise) {
    sessionBootstrapPromise = fetch(API_SESSION_URL, {
      method: "GET",
      credentials: "include",
    }).then((res) => {
      if (!res.ok) {
        throw new Error(`API session failed (${res.status})`);
      }
    });
  }

  return sessionBootstrapPromise.catch((err) => {
    sessionBootstrapPromise = null;
    throw err;
  });
}

/** Clears the cached bootstrap promise (for tests or after auth errors). */
export function resetApiSessionCache() {
  sessionBootstrapPromise = null;
}
