import {
  API_SESSION_URL,
  STATUS_ONLY_QUERY_PARAM,
  TURNSTILE_SITE_KEY,
  TURNSTILE_TOKEN_QUERY_PARAM,
} from "../constants";
import { isTurnstileEnabled, requestTurnstileToken } from "./turnstile";

export const DEFAULT_MAINTENANCE_MESSAGE =
  "Search is temporarily unavailable. Please try again later.";

const MAINTENANCE_MODE_HEADER = "X-Maintenance-Mode";
const MAINTENANCE_MESSAGE_HEADER = "X-Maintenance-Message";
const NOTICE_MESSAGE_HEADER = "X-Notice-Message";

/**
 * @typedef {{
 *   maintenanceMode: boolean,
 *   maintenanceMessage: string,
 *   noticeMessage: string,
 * }} SiteStatus
 */

/**
 * @typedef {SiteStatus & {
 *   sessionMintDurationMs: number,
 * }} SessionBootstrapTiming
 */

/** Refresh before default API session TTL (15m) so idle tabs stay authorized. */
export const API_SESSION_REFRESH_INTERVAL_MS = 10 * 60 * 1000;

/** Network blips are common in private browsing; retry session mint. */
const SESSION_MINT_MAX_ATTEMPTS = 3;

let sessionBootstrapPromise = null;
/** @type {SessionBootstrapTiming | null} */
let cachedSessionBootstrap = null;
let siteStatusPromise = null;
/** @type {SiteStatus | null} */
let cachedSiteStatus = null;

/** Resolved bootstrap timing when session mint has completed (for fast UI hydration). */
export function getCachedSessionBootstrap() {
  return cachedSessionBootstrap;
}

/** Resolved site status when the status-only probe has completed. */
export function getCachedSiteStatus() {
  return cachedSiteStatus;
}

/**
 * Fetches notice/maintenance without Turnstile or session cookies so banners can
 * render before session bootstrap completes.
 *
 * @returns {Promise<SiteStatus>}
 */
export async function fetchSiteStatus(options = {}) {
  const { forceRefresh = false } = options;

  if (forceRefresh) {
    siteStatusPromise = null;
  }

  if (!siteStatusPromise) {
    siteStatusPromise = loadSiteStatus().catch((err) => {
      siteStatusPromise = null;
      throw err;
    });
  }

  const status = await siteStatusPromise;
  cachedSiteStatus = status;
  return status;
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

export function parseNoticeFromSessionResponse(res) {
  const message = res.headers.get(NOTICE_MESSAGE_HEADER);
  return {
    noticeMessage:
      typeof message === "string" && message.trim() !== ""
        ? message.trim()
        : "",
  };
}

export function parseSiteStatusFromSessionResponse(res) {
  return {
    ...parseMaintenanceFromSessionResponse(res),
    ...parseNoticeFromSessionResponse(res),
  };
}

export function parseSiteStatusFromSessionBody(body) {
  const maintenanceMode =
    typeof body?.maintenanceMode === "boolean"
      ? body.maintenanceMode
      : typeof body?.maintenanceMode === "string"
        ? body.maintenanceMode.toLowerCase() === "true"
        : Boolean(body?.maintenanceMode);
  const maintenanceMessage =
    typeof body?.maintenanceMessage === "string" &&
    body.maintenanceMessage.trim() !== ""
      ? body.maintenanceMessage.trim()
      : maintenanceMode
        ? DEFAULT_MAINTENANCE_MESSAGE
        : "";
  const noticeMessage =
    typeof body?.noticeMessage === "string" && body.noticeMessage.trim() !== ""
      ? body.noticeMessage.trim()
      : "";

  return {
    maintenanceMode,
    maintenanceMessage,
    noticeMessage,
  };
}

export async function parseSiteStatusFromSession(res) {
  const fromHeaders = parseSiteStatusFromSessionResponse(res);
  const bodyText = await res.text();
  if (bodyText.trim() === "") {
    return fromHeaders;
  }

  try {
    const body = JSON.parse(bodyText);
    const fromBody = parseSiteStatusFromSessionBody(body);
    return {
      maintenanceMode: fromBody.maintenanceMode || fromHeaders.maintenanceMode,
      maintenanceMessage:
        fromBody.maintenanceMessage || fromHeaders.maintenanceMessage,
      noticeMessage: fromBody.noticeMessage || fromHeaders.noticeMessage,
    };
  } catch {
    return fromHeaders;
  }
}

function joinedBootstrapTiming(bootstrapTiming) {
  return {
    ...bootstrapTiming,
    sessionMintDurationMs: 0,
  };
}

/**
 * Ensures the browser holds a valid HttpOnly API session cookie (minted by GET /session on the API host).
 * Safe to call repeatedly; concurrent callers share one in-flight request unless forceRefresh.
 *
 * @returns {Promise<SessionBootstrapTiming>} Time spent in this call. Cached sessions return zeros.
 */
export async function ensureApiSession(options = {}) {
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
    cachedSessionBootstrap = bootstrapTiming;
    if (!initiatedBootstrap) {
      return joinedBootstrapTiming(bootstrapTiming);
    }
    return bootstrapTiming;
  } catch (err) {
    sessionBootstrapPromise = null;
    cachedSessionBootstrap = null;
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
      return {
        sessionMintDurationMs,
        maintenanceMode: timing.maintenanceMode,
        maintenanceMessage: timing.maintenanceMessage,
        noticeMessage: timing.noticeMessage,
      };
    } catch (err) {
      lastError = err;
    }
  }

  throw lastError ?? new Error("API session failed");
}

async function loadSiteStatus() {
  const params = new URLSearchParams({
    [STATUS_ONLY_QUERY_PARAM]: "1",
  });
  const res = await fetch(`${API_SESSION_URL}?${params.toString()}`, {
    method: "GET",
    credentials: "omit",
  });

  if (!res.ok) {
    throw new Error(`Site status failed (${res.status})`);
  }

  return parseSiteStatusFromSession(res);
}

async function mintApiSession() {
  // Network failures surface as TypeError; rethrow as-is so the UI keeps the
  // accurate "unable to connect" copy instead of blaming session verification.
  const sessionMintStart = performance.now();
  const fetchOptions = {
    method: "GET",
    credentials: "include",
  };

  let sessionUrl = API_SESSION_URL;
  if (isTurnstileEnabled(TURNSTILE_SITE_KEY)) {
    const turnstileToken = await requestTurnstileToken(TURNSTILE_SITE_KEY);
    const params = new URLSearchParams({
      [TURNSTILE_TOKEN_QUERY_PARAM]: turnstileToken,
    });
    // Query param avoids CORS preflight for X-Turnstile-Token (API Gateway OPTIONS
    // does not allow that header today). TODO(api-abuse): POST /session when available.
    sessionUrl = `${API_SESSION_URL}?${params.toString()}`;
  }

  const res = await fetch(sessionUrl, fetchOptions);
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

  const siteStatus = await parseSiteStatusFromSession(res);
  return {
    sessionMintDurationMs,
    ...siteStatus,
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
  cachedSessionBootstrap = null;
}

/** Clears cached site status (for tests). */
export function resetSiteStatusCache() {
  siteStatusPromise = null;
  cachedSiteStatus = null;
}

/** User-facing copy when the initial session bootstrap fails. */
export function formatSessionBootstrapError(err) {
  if (err instanceof Error && err.message) {
    const message = err.message.trim();
    if (message.toLowerCase().includes("turnstile")) {
      return "Session verification failed. Please refresh the page and try again.";
    }
    if (message.toLowerCase().includes("api session")) {
      return message;
    }
    return `Unable to start a search session: ${message}`;
  }
  return "Unable to start a search session. Please refresh and try again.";
}
