const TURNSTILE_SCRIPT_SRC =
  "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";

const TURNSTILE_TIMEOUT_MS = 30_000;

let scriptLoadPromise = null;
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
    const existing = document.querySelector(
      `script[src^="https://challenges.cloudflare.com/turnstile/"]`,
    );
    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener(
        "error",
        () => reject(new Error("turnstile script failed")),
        { once: true },
      );
      return;
    }

    const script = document.createElement("script");
    script.src = TURNSTILE_SCRIPT_SRC;
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("turnstile script failed"));
    document.head.appendChild(script);
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
  if (pendingTokenPromise) {
    pendingTokenPromise.resolve(token);
    pendingTokenPromise = null;
  }
}

export function onTurnstileExpired() {
  cachedToken = "";
}

export function onTurnstileError() {
  cachedToken = "";
  rejectPendingTokenWaiters(new Error("turnstile challenge failed"));
}

export function teardownTurnstileWidget() {
  rejectPendingTokenWaiters(new Error("turnstile widget teardown"));
  cachedToken = "";
  if (widgetId != null && window.turnstile?.remove) {
    window.turnstile.remove(widgetId);
  }
  widgetId = null;
  widgetContainer = null;
  widgetReady = createDeferred();
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
    const id = window.turnstile.render(container, {
      sitekey: siteKey(),
      size: "invisible",
      callback: onTurnstileToken,
      "expired-callback": onTurnstileExpired,
      "error-callback": onTurnstileError,
    });
    widgetContainer = container;
    registerTurnstileWidget(id);
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

export async function obtainTurnstileToken() {
  if (!isTurnstileConfigured()) {
    return "";
  }

  await widgetReady.promise;
  if (widgetId == null) {
    throw new Error("turnstile widget not ready");
  }

  if (cachedToken) {
    return cachedToken;
  }

  window.turnstile.execute(widgetId);
  return waitForTurnstileToken();
}

export function resetTurnstileChallenge() {
  cachedToken = "";
  rejectPendingTokenWaiters(new Error("turnstile challenge reset"));
  if (widgetId != null && window.turnstile) {
    window.turnstile.reset(widgetId);
  }
}
