import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
import { TURNSTILE_SITE_KEY } from "./constants";
import { ensureApiSession, fetchSiteStatus } from "./utils/apiSession";
import { isTurnstileEnabled, preloadTurnstile } from "./utils/turnstile";

// Fetch notice/maintenance immediately (no Turnstile). Overlap session mint with hydration.
void fetchSiteStatus();
if (isTurnstileEnabled(TURNSTILE_SITE_KEY)) {
  void preloadTurnstile(TURNSTILE_SITE_KEY);
}
void ensureApiSession();

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
