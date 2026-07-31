const TURNSTILE_SCRIPT_SRC =
  "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";

const TURNSTILE_TIMEOUT_MS = 30_000;
/** Caps wait when the widget never registers (e.g. script blocked by an extension). */
const TURNSTILE_WIDGET_READY_TIMEOUT_MS = 12_000;
const TURNSTILE_SCRIPT_LOAD_TIMEOUT_MS = 8_000;

let scriptLoadPromise = null;
let widgetPreparePromise = null;
let widgetId = null;
let widgetContainer = null;

function createDeferred() {
  let resolve;
  let reject;
  const promise = new Promise((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

let widgetReady = createDeferred();

let cachedToken = "";
let pendingTokenPromise = null;
/** True after execute() until token, error, expiry, reset, or teardown. */
let challengeInFlight = false;

function siteKey() {
  const key = import.meta.env.VITE_TURNSTILE_SITE_KEY;
  return typeof key === "string" ? key.trim() : "";
}

export function isTurnstileConfigured() {
  return siteKey() !== "";
}

function loadTurnstileScript() {
  if (!isTurnstileConfigured()) {
    return Promise.resolve();
  }
  if (window.turnstile) {
    return Promise.resolve();
  }
  if (scriptLoadPromise) {
    return scriptLoadPromise;
  }

  scriptLoadPromise = new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error("turnstile script timeout"));
    }, TURNSTILE_SCRIPT_LOAD_TIMEOUT_MS);

    const finish = (handler) => {
      clearTimeout(timeout);
      handler();
    };

    const existing = document.querySelector(
      `script[src^="https://challenges.cloudflare.com/turnstile/"]`,
    );
    if (existing) {
      if (window.turnstile) {
        finish(resolve);
        return;
      }
      existing.addEventListener("load", () => finish(resolve), { once: true });
      existing.addEventListener(
        "error",
        () => finish(() => reject(new Error("turnstile script failed"))),
        { once: true },
      );
      return;
    }

    const script = document.createElement("script");
    script.src = TURNSTILE_SCRIPT_SRC;
    script.async = true;
    script.defer = true;
    script.onload = () => finish(resolve);
    script.onerror = () =>
      finish(() => reject(new Error("turnstile script failed")));
    document.head.appendChild(script);
  });

  // A rejected load must not be cached — prepare retries need a fresh attempt.
  scriptLoadPromise.catch(() => {
    scriptLoadPromise = null;
  });

  return scriptLoadPromise;
}

function rejectPendingTokenWaiters(err) {
  if (pendingTokenPromise) {
    pendingTokenPromise.reject(err);
    pendingTokenPromise = null;
  }
}

function failWidgetReady(err) {
  widgetReady.reject(
    err instanceof Error ? err : new Error("turnstile widget failed"),
  );
  widgetReady = createDeferred();
}

export function registerTurnstileWidget(id) {
  widgetId = id;
  widgetReady.resolve();
}

export function onTurnstileToken(token) {
  cachedToken = token;
  challengeInFlight = false;
  if (pendingTokenPromise) {
    pendingTokenPromise.resolve(token);
    pendingTokenPromise = null;
  }
}

export function onTurnstileExpired() {
  cachedToken = "";
  challengeInFlight = false;
  startTurnstileChallenge();
}

export function onTurnstileError() {
  cachedToken = "";
  challengeInFlight = false;
  rejectPendingTokenWaiters(new Error("turnstile challenge failed"));
}

export function teardownTurnstileWidget() {
  rejectPendingTokenWaiters(new Error("turnstile widget teardown"));
  cachedToken = "";
  challengeInFlight = false;
  widgetPreparePromise = null;
  if (widgetId != null && window.turnstile?.remove) {
    window.turnstile.remove(widgetId);
  }
  widgetId = null;
  widgetContainer = null;
  failWidgetReady(new Error("turnstile widget teardown"));
}

/**
 * Starts widget preparation once; concurrent callers share the same promise.
 * TurnstileBootstrap should call this as early as possible on mount.
 */
export function beginTurnstileWidgetPrepare(container, options = {}) {
  if (!isTurnstileConfigured() || !container) {
    return Promise.resolve();
  }

  if (widgetPreparePromise) {
    return widgetPreparePromise;
  }

  widgetPreparePromise = prepareTurnstileWidget(container, options).catch(
    (err) => {
      widgetPreparePromise = null;
      throw err;
    },
  );
  return widgetPreparePromise;
}

export async function prepareTurnstileWidget(container, options = {}) {
  const { signal } = options;
  if (!isTurnstileConfigured() || !container) {
    return;
  }

  try {
    await loadTurnstileScript();
  } catch (err) {
    failWidgetReady(err);
    throw err;
  }

  if (signal?.aborted || !container.isConnected) {
    return;
  }

  if (widgetId != null) {
    if (widgetContainer === container && container.isConnected) {
      startTurnstileChallenge();
      return;
    }
    teardownTurnstileWidget();
  }

  if (!window.turnstile) {
    const err = new Error("turnstile script unavailable");
    failWidgetReady(err);
    throw err;
  }

  try {
    // Widget mode (Invisible / Managed / …) is chosen in the Cloudflare dashboard for the
    // sitekey. execution: "execute" defers the challenge until startTurnstileChallenge().
    const id = window.turnstile.render(container, {
      sitekey: siteKey(),
      execution: "execute",
      callback: onTurnstileToken,
      "expired-callback": onTurnstileExpired,
      "error-callback": onTurnstileError,
    });
    widgetContainer = container;
    registerTurnstileWidget(id);
    startTurnstileChallenge();
  } catch (err) {
    failWidgetReady(err);
    throw err;
  }
}

function waitForTurnstileToken() {
  if (cachedToken) {
    return Promise.resolve(cachedToken);
  }

  if (pendingTokenPromise) {
    return pendingTokenPromise.promise;
  }

  let resolve;
  let reject;
  const promise = new Promise((res, rej) => {
    resolve = res;
    reject = rej;
  });
  pendingTokenPromise = { promise, resolve, reject };

  const timeout = setTimeout(() => {
    if (pendingTokenPromise) {
      pendingTokenPromise.reject(new Error("turnstile timeout"));
      pendingTokenPromise = null;
    }
  }, TURNSTILE_TIMEOUT_MS);

  return promise.finally(() => clearTimeout(timeout));
}

function waitForWidgetReady() {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      reject(new Error("turnstile widget timeout"));
    }, TURNSTILE_WIDGET_READY_TIMEOUT_MS);

    widgetReady.promise.then(
      (value) => {
        clearTimeout(timeout);
        resolve(value);
      },
      (err) => {
        clearTimeout(timeout);
        reject(err);
      },
    );
  });
}

async function ensureTurnstileWidgetReady() {
  if (widgetPreparePromise) {
    await widgetPreparePromise;
  }
  await waitForWidgetReady();
}

function startTurnstileChallenge() {
  if (widgetId == null || !window.turnstile?.execute) {
    return;
  }
  if (cachedToken || challengeInFlight) {
    return;
  }
  challengeInFlight = true;
  window.turnstile.execute(widgetId);
}

export async function obtainTurnstileToken() {
  if (!isTurnstileConfigured()) {
    return "";
  }

  await ensureTurnstileWidgetReady();
  if (widgetId == null || !window.turnstile) {
    throw new Error("turnstile widget not ready");
  }

  if (cachedToken) {
    return cachedToken;
  }

  startTurnstileChallenge();
  return waitForTurnstileToken();
}

export function resetTurnstileChallenge() {
  cachedToken = "";
  challengeInFlight = false;
  rejectPendingTokenWaiters(new Error("turnstile challenge reset"));
  if (widgetId != null && window.turnstile) {
    window.turnstile.reset(widgetId);
  }
  startTurnstileChallenge();
}
