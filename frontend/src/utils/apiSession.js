import { API_SESSION_URL } from "../constants";
import {
  obtainTurnstileToken,
  resetTurnstileChallenge,
} from "./turnstileSession";

/** Refresh before default API session TTL (15m) so idle tabs stay authorized. */
export const API_SESSION_REFRESH_INTERVAL_MS = 10 * 60 * 1000;

/** Turnstile / network blips are common in private browsing; retry with a fresh token. */
const SESSION_MINT_MAX_ATTEMPTS = 3;

/** Treat near-instant awaits as a warm (already-minted) session. */
const WARM_SESSION_WAIT_MS = 2;

let sessionBootstrapPromise = null;

function nowMs() {
  return typeof performance !== "undefined" &&
    typeof performance.now === "function"
    ? performance.now()
    : Date.now();
}

/**
 * @typedef {{
 *   turnstileMs: number | null,
 *   sessionFetchMs: number | null,
 *   totalMs: number,
 *   attempts: number,
 *   reused: boolean,
 *   joinedInFlight: boolean,
 * }} SessionTiming
 */

/** @returns {SessionTiming} */
export function emptySessionTiming() {
  return {
    turnstileMs: 0,
    sessionFetchMs: 0,
    totalMs: 0,
    attempts: 0,
    reused: true,
    joinedInFlight: false,
  };
}

/**
 * Ensures the browser holds a valid HttpOnly API session cookie (minted by GET /session on the API host).
 * Safe to call repeatedly; concurrent callers share one in-flight request unless forceRefresh.
 *
 * Returns client-side timing for Turnstile + session mint so SearchStats can break down waits
 * that sit outside API `totalDurationMs`.
 *
 * @returns {Promise<SessionTiming>}
 */
export async function ensureApiSession(options = {}) {
  const { forceRefresh = false } = options;

  if (forceRefresh) {
    sessionBootstrapPromise = null;
    resetTurnstileChallenge();
  }

  const callStarted = nowMs();
  const startedMint = !sessionBootstrapPromise;

  if (!sessionBootstrapPromise) {
    sessionBootstrapPromise = bootstrapSessionWithRetry();
  }

  try {
    const mintTiming = await sessionBootstrapPromise;
    const waitedMs = Math.max(0, Math.round(nowMs() - callStarted));

    // Warm cookie / already-resolved bootstrap: do not attribute prior mint cost to this call.
    if (waitedMs <= WARM_SESSION_WAIT_MS) {
      return emptySessionTiming();
    }

    if (startedMint) {
      return {
        turnstileMs: mintTiming.turnstileMs,
        sessionFetchMs: mintTiming.sessionFetchMs,
        totalMs: mintTiming.totalMs,
        attempts: mintTiming.attempts,
        reused: false,
        joinedInFlight: false,
      };
    }

    // Joined an in-flight bootstrap (e.g. mount refresh still running): report wait only.
    return {
      turnstileMs: null,
      sessionFetchMs: null,
      totalMs: waitedMs,
      attempts: mintTiming.attempts,
      reused: false,
      joinedInFlight: true,
    };
  } catch (err) {
    sessionBootstrapPromise = null;
    throw err;
  }
}

async function bootstrapSessionWithRetry() {
  let lastError = null;
  let turnstileMs = 0;
  let sessionFetchMs = 0;

  for (let attempt = 1; attempt <= SESSION_MINT_MAX_ATTEMPTS; attempt += 1) {
    const attemptTiming = await mintApiSession();
    turnstileMs += attemptTiming.turnstileMs;
    sessionFetchMs += attemptTiming.sessionFetchMs;

    if (attemptTiming.ok) {
      return {
        turnstileMs,
        sessionFetchMs,
        totalMs: turnstileMs + sessionFetchMs,
        attempts: attempt,
      };
    }

    lastError = attemptTiming.error;
    // Drop any pending/used token so the next attempt runs a fresh challenge.
    resetTurnstileChallenge();
  }

  throw lastError ?? new Error("API session failed");
}

/**
 * @returns {Promise<{
 *   ok: boolean,
 *   turnstileMs: number,
 *   sessionFetchMs: number,
 *   error?: Error,
 * }>}
 */
async function mintApiSession() {
  const headers = {};
  let turnstileMs = 0;
  let sessionFetchMs = 0;

  const turnstileStarted = nowMs();
  try {
    const turnstileToken = await obtainTurnstileToken();
    turnstileMs = Math.max(0, Math.round(nowMs() - turnstileStarted));
    if (turnstileToken) {
      headers["CF-Turnstile-Response"] = turnstileToken;
    }
  } catch (err) {
    turnstileMs = Math.max(0, Math.round(nowMs() - turnstileStarted));
    return {
      ok: false,
      turnstileMs,
      sessionFetchMs: 0,
      error: err instanceof Error ? err : new Error("turnstile failed"),
    };
  }

  // Network failures surface as TypeError; rethrow as-is so the UI keeps the
  // accurate "unable to connect" copy instead of blaming session verification.
  const fetchStarted = nowMs();
  let res;
  try {
    res = await fetch(API_SESSION_URL, {
      method: "GET",
      credentials: "include",
      headers,
    });
  } catch (err) {
    sessionFetchMs = Math.max(0, Math.round(nowMs() - fetchStarted));
    // Tokens are single-use; never reuse after a mint attempt.
    resetTurnstileChallenge();
    return {
      ok: false,
      turnstileMs,
      sessionFetchMs,
      error: err instanceof Error ? err : new Error("API session failed"),
    };
  }

  sessionFetchMs = Math.max(0, Math.round(nowMs() - fetchStarted));
  resetTurnstileChallenge();

  if (!res.ok) {
    let error;
    if (res.status === 403) {
      error = new Error(
        "API session verification failed. Please try searching again.",
      );
    } else {
      error = new Error(`API session failed (${res.status})`);
    }
    return { ok: false, turnstileMs, sessionFetchMs, error };
  }

  return { ok: true, turnstileMs, sessionFetchMs };
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
