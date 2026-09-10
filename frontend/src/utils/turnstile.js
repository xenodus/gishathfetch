const TURNSTILE_SCRIPT_URL =
  "https://challenges.cloudflare.com/turnstile/v0/api.js";
const TURNSTILE_TIMEOUT_MS = 30_000;

let scriptLoadPromise = null;
let widgetId = null;
let widgetContainer = null;
/** @type {{ resolve: (token: string) => void, reject: (err: Error) => void } | null} */
let pendingToken = null;

export function isTurnstileEnabled(siteKey) {
  return typeof siteKey === "string" && siteKey.trim() !== "";
}

function loadTurnstileScript() {
  if (typeof window !== "undefined" && window.turnstile) {
    return Promise.resolve();
  }
  if (!scriptLoadPromise) {
    scriptLoadPromise = new Promise((resolve, reject) => {
      const script = document.createElement("script");
      script.src = TURNSTILE_SCRIPT_URL;
      script.async = true;
      script.defer = true;
      script.onload = () => resolve();
      script.onerror = () => {
        scriptLoadPromise = null;
        reject(new Error("Turnstile failed to load"));
      };
      document.head.appendChild(script);
    });
  }
  return scriptLoadPromise;
}

function ensureWidget(siteKey) {
  if (widgetId !== null) {
    return widgetId;
  }

  widgetContainer = document.createElement("div");
  widgetContainer.style.display = "none";
  widgetContainer.setAttribute("aria-hidden", "true");
  document.body.appendChild(widgetContainer);

  widgetId = window.turnstile.render(widgetContainer, {
    sitekey: siteKey,
    size: "invisible",
    callback: (token) => {
      pendingToken?.resolve(token);
    },
    "error-callback": () => {
      pendingToken?.reject(new Error("Turnstile verification failed"));
    },
    "expired-callback": () => {
      pendingToken?.reject(new Error("Turnstile token expired"));
    },
  });

  return widgetId;
}

/**
 * Runs invisible Turnstile and returns a one-time token for POST /session.
 * Safe to call on every session mint and background refresh.
 */
export async function requestTurnstileToken(siteKey) {
  if (!isTurnstileEnabled(siteKey)) {
    throw new Error("Turnstile site key is not configured");
  }

  await loadTurnstileScript();
  const id = ensureWidget(siteKey.trim());

  if (pendingToken) {
    throw new Error("Turnstile request already in progress");
  }

  return await new Promise((resolve, reject) => {
    const timeout = window.setTimeout(() => {
      if (!pendingToken) {
        return;
      }
      pendingToken = null;
      reject(new Error("Turnstile timed out"));
    }, TURNSTILE_TIMEOUT_MS);

    pendingToken = {
      resolve: (token) => {
        window.clearTimeout(timeout);
        pendingToken = null;
        resolve(token);
      },
      reject: (err) => {
        window.clearTimeout(timeout);
        pendingToken = null;
        reject(err);
      },
    };

    window.turnstile.reset(id);
    window.turnstile.execute(id);
  });
}

/** Clears widget state (tests only). */
export function resetTurnstileWidget() {
  if (
    widgetId !== null &&
    typeof window !== "undefined" &&
    window.turnstile?.remove
  ) {
    window.turnstile.remove(widgetId);
  }
  widgetId = null;
  widgetContainer?.remove();
  widgetContainer = null;
  pendingToken = null;
  scriptLoadPromise = null;
}
