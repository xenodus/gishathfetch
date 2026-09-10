import assert from "node:assert/strict";
import { isTurnstileEnabled } from "./turnstile.js";

assert.equal(isTurnstileEnabled(""), false);
assert.equal(isTurnstileEnabled("   "), false);
assert.equal(isTurnstileEnabled("1x00000000000000000000AA"), true);
