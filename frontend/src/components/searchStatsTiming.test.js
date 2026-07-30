import assert from "node:assert/strict";

/**
 * Mirrors SearchStats sessionAccountedMs so the accounting stays covered without a DOM test runner.
 * @param {{
 *   turnstileMs?: number | null,
 *   sessionFetchMs?: number | null,
 *   totalMs?: number,
 *   reused?: boolean,
 *   joinedInFlight?: boolean,
 * } | null} sessionTiming
 */
function sessionAccountedMs(sessionTiming) {
  if (!sessionTiming || sessionTiming.reused) {
    return 0;
  }
  if (sessionTiming.joinedInFlight) {
    return Number.isFinite(sessionTiming.totalMs) ? sessionTiming.totalMs : 0;
  }
  const turnstile =
    Number.isFinite(sessionTiming.turnstileMs) && sessionTiming.turnstileMs > 0
      ? sessionTiming.turnstileMs
      : 0;
  const sessionFetch =
    Number.isFinite(sessionTiming.sessionFetchMs) &&
    sessionTiming.sessionFetchMs > 0
      ? sessionTiming.sessionFetchMs
      : 0;
  return turnstile + sessionFetch;
}

assert.equal(sessionAccountedMs(null), 0);
assert.equal(
  sessionAccountedMs({
    reused: true,
    turnstileMs: 5000,
    sessionFetchMs: 200,
    totalMs: 5200,
  }),
  0,
);
assert.equal(
  sessionAccountedMs({
    reused: false,
    joinedInFlight: true,
    turnstileMs: null,
    sessionFetchMs: null,
    totalMs: 1200,
  }),
  1200,
);
assert.equal(
  sessionAccountedMs({
    reused: false,
    joinedInFlight: false,
    turnstileMs: 8000,
    sessionFetchMs: 350,
    totalMs: 8350,
  }),
  8350,
);
assert.equal(
  sessionAccountedMs({
    reused: false,
    joinedInFlight: false,
    turnstileMs: 0,
    sessionFetchMs: 120,
    totalMs: 120,
  }),
  120,
);

const clientDurationMs = 12400;
const totalDurationMs = 1000;
const accounted = sessionAccountedMs({
  reused: false,
  joinedInFlight: false,
  turnstileMs: 8200,
  sessionFetchMs: 350,
  totalMs: 8550,
});
const otherMs = Math.max(0, clientDurationMs - accounted - totalDurationMs);
assert.equal(otherMs, 2850);

console.log("searchStatsTiming.test.js: ok");
