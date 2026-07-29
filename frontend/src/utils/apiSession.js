import { API_SESSION_URL } from "../constants";
import {
  obtainTurnstileToken,
  resetTurnstileChallenge,
} from "./turnstileSession";

/** Refresh before default API session TTL (15m) so idle tabs stay authorized. */
export const API_SESSION_REFRESH_INTERVAL_MS = 10 * 60 * 1000;

/** Turnstile / network blips are common in private browsing; retry with a fresh token. */
const SESSION_MINT_MAX_ATTEMPTS = 3;

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
    sessionBootstrapPromise = bootstrapSessionWithRetry();
  }

  try {
    await sessionBootstrapPromise;
  } catch (err) {
    sessionBootstrapPromise = null;
    throw err;
  }
}

async function bootstrapSessionWithRetry() {
  let lastError = null;

  for (let attempt = 1; attempt <= SESSION_MINT_MAX_ATTEMPTS; attempt += 1) {
    try {
      await mintApiSession();
      return;
    } catch (err) {
      lastError = err;
      // Drop any pending/used token so the next attempt runs a fresh challenge.
      resetTurnstileChallenge();
    }
  }

  throw lastError ?? new Error("API session failed");
}

async function mintApiSession() {
  const headers = {};
  const turnstileToken = await obtainTurnstileToken();
  if (turnstileToken) {
    headers["CF-Turnstile-Response"] = turnstileToken;
  }

  // Network failures surface as TypeError; rethrow as-is so the UI keeps the
  // accurate "unable to connect" copy instead of blaming session verification.
  const res = await fetch(API_SESSION_URL, {
    method: "GET",
    credentials: "include",
    headers,
  });

  // Tokens are single-use; never reuse after a mint attempt.
  resetTurnstileChallenge();

  if (!res.ok) {
    if (res.status === 403) {
      throw new Error(
        "API session verification failed. Please try searching again.",
      );
    }
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
