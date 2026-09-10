import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.jsx";
import { TURNSTILE_SITE_KEY } from "./constants";
import { ensureApiSession } from "./utils/apiSession";
import { isTurnstileEnabled, preloadTurnstile } from "./utils/turnstile";

// Overlap Turnstile + session mint with React hydration so trending ?s= links
// show the API notice and start searching sooner after landing.
if (isTurnstileEnabled(TURNSTILE_SITE_KEY)) {
  void preloadTurnstile(TURNSTILE_SITE_KEY);
}
void ensureApiSession();

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
