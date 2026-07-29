import { API_SESSION_URL } from "../constants";
import {
  obtainTurnstileToken,
  resetTurnstileChallenge,
} from "./turnstileSession";

/** Refresh before default API session TTL (15m) so idle tabs stay authorized. */
export const API_SESSION_REFRESH_INTERVAL_MS = 10 * 60 * 1000;

let sessionBootstrapPromise = null;

/**
 * Ensures the browser holds a valid HttpOnly API session cookie (minted by GET /session on the API host).
 * Safe to call repeatedly; concurrent callers share one in-flight request unless forceRefresh.
 */
export async function ensureApiSession(options = {}) {
  const { forceRefresh = false } = options;

  if (forceRefresh) {
    sessionBootstrapPromise = null;
    resetTurnstileChallenge();
  }

  if (!sessionBootstrapPromise) {
    sessionBootstrapPromise = bootstrapSession();
  }

  try {
    await sessionBootstrapPromise;
  } catch (err) {
    sessionBootstrapPromise = null;
    throw err;
  }
}

async function bootstrapSession() {
  const headers = {};
  const turnstileToken = await obtainTurnstileToken();
  if (turnstileToken) {
    headers["CF-Turnstile-Response"] = turnstileToken;
  }

  const res = await fetch(API_SESSION_URL, {
    method: "GET",
    credentials: "include",
    headers,
  });

  if (!res.ok) {
    resetTurnstileChallenge();
    throw new Error(`API session failed (${res.status})`);
  }
}

/** True when search API rejected the request for missing or expired session cookie. */
export function isApiSessionAccessDenied(message, statusCode) {
  if (statusCode !== 403) {
    return false;
  }
  if (typeof message !== "string" || !message) {
    return false;
  }
  const lower = message.toLowerCase();
  return lower === "session required" || lower === "session expired";
}

/** Clears the cached bootstrap promise (for tests or after auth errors). */
export function resetApiSessionCache() {
  sessionBootstrapPromise = null;
  resetTurnstileChallenge();
}
