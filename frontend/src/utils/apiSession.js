import { API_SESSION_URL } from "../constants";
import {
  obtainTurnstileToken,
  resetTurnstileChallenge,
} from "./turnstileSession";

let sessionBootstrapPromise = null;

/**
 * Ensures the browser holds a valid HttpOnly API session cookie (minted by GET /session on the API host).
 * Safe to call repeatedly; concurrent callers share one in-flight request.
 */
export async function ensureApiSession() {
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

/** Clears the cached bootstrap promise (for tests or after auth errors). */
export function resetApiSessionCache() {
  sessionBootstrapPromise = null;
  resetTurnstileChallenge();
}
