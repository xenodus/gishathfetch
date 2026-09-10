import assert from "node:assert/strict";
import {
  isTurnstileEnabled,
  preloadTurnstile,
  TURNSTILE_SCRIPT_URL,
} from "./turnstile.js";

assert.equal(isTurnstileEnabled(""), false);
assert.equal(isTurnstileEnabled("   "), false);
assert.equal(isTurnstileEnabled("1x00000000000000000000AA"), true);
assert.equal(
  TURNSTILE_SCRIPT_URL,
  "https://challenges.cloudflare.com/turnstile/v0/api.js",
);

await preloadTurnstile("");
await preloadTurnstile("   ");
