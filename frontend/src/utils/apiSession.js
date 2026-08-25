import { API_SESSION_URL } from "../constants";

export const DEFAULT_MAINTENANCE_MESSAGE =
  "Search is temporarily unavailable. Please try again later.";

const MAINTENANCE_MODE_HEADER = "X-Maintenance-Mode";
const MAINTENANCE_MESSAGE_HEADER = "X-Maintenance-Message";

/**
 * @typedef {{
 *   sessionMintDurationMs: number,
 *   maintenanceMode?: boolean,
 *   maintenanceMessage?: string,
 * }} SessionBootstrapTiming
 */

/** @type {SessionBootstrapTiming} */
function noSessionWork() {
  return {
    sessionMintDurationMs: 0,
    maintenanceMode: false,
    maintenanceMessage: "",
  };
}

export function parseMaintenanceFromSessionResponse(res) {
  if (res.headers.get(MAINTENANCE_MODE_HEADER) !== "1") {
    return {
      maintenanceMode: false,
      maintenanceMessage: "",
    };
  }

  const message = res.headers.get(MAINTENANCE_MESSAGE_HEADER);
  return {
    maintenanceMode: true,
    maintenanceMessage:
      typeof message === "string" && message.trim() !== ""
        ? message
        : DEFAULT_MAINTENANCE_MESSAGE,
  };
}

function attributeJoinedBootstrapWait(totalWaitMs, bootstrapTiming) {
  const mintMs = bootstrapTiming.sessionMintDurationMs ?? 0;
  return {
    sessionMintDurationMs: mintMs > 0 ? mintMs : totalWaitMs,
  };
}

/** Refresh before default API session TTL (15m) so idle tabs stay authorized. */
export const API_SESSION_REFRESH_INTERVAL_MS = 10 * 60 * 1000;

/** Network blips are common in private browsing; retry session mint. */
const SESSION_MINT_MAX_ATTEMPTS = 3;

let sessionBootstrapPromise = null;

/**
 * Ensures the browser holds a valid HttpOnly API session cookie (minted by GET /session on the API host).
 * Safe to call repeatedly; concurrent callers share one in-flight request unless forceRefresh.
 *
 * @returns {Promise<SessionBootstrapTiming>} Time spent in this call. Cached sessions return zeros.
 */
export async function ensureApiSession(options = {}) {
  const waitStart = performance.now();
  const { forceRefresh = false } = options;

  if (forceRefresh) {
    sessionBootstrapPromise = null;
  }

  const initiatedBootstrap = !sessionBootstrapPromise;
  if (!sessionBootstrapPromise) {
    sessionBootstrapPromise = bootstrapSessionWithRetry();
  }

  try {
    const bootstrapTiming = await sessionBootstrapPromise;
    const totalWaitMs = Math.round(performance.now() - waitStart);

    if (!initiatedBootstrap && totalWaitMs > 0) {
      return attributeJoinedBootstrapWait(totalWaitMs, bootstrapTiming);
    }

    if (!initiatedBootstrap) {
      return noSessionWork();
    }

    return bootstrapTiming;
  } catch (err) {
    sessionBootstrapPromise = null;
    throw err;
  }
}

async function bootstrapSessionWithRetry() {
  let lastError = null;
  let sessionMintDurationMs = 0;

  for (let attempt = 1; attempt <= SESSION_MINT_MAX_ATTEMPTS; attempt += 1) {
    try {
      const timing = await mintApiSession();
      sessionMintDurationMs += timing.sessionMintDurationMs;
      return { sessionMintDurationMs };
    } catch (err) {
      lastError = err;
    }
  }

  throw lastError ?? new Error("API session failed");
}

async function mintApiSession() {
  // Network failures surface as TypeError; rethrow as-is so the UI keeps the
  // accurate "unable to connect" copy instead of blaming session verification.
  const sessionMintStart = performance.now();
  const res = await fetch(API_SESSION_URL, {
    method: "GET",
    credentials: "include",
  });
  const sessionMintDurationMs = Math.round(
    performance.now() - sessionMintStart,
  );

  if (!res.ok) {
    if (res.status === 403) {
      throw new Error(
        "API session verification failed. Please try searching again.",
      );
    }
    throw new Error(`API session failed (${res.status})`);
  }

  const maintenance = parseMaintenanceFromSessionResponse(res);
  return {
    sessionMintDurationMs,
    ...maintenance,
  };
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
}
