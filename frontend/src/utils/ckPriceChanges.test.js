import assert from "node:assert/strict";
import test from "node:test";
import {
  formatPriceChangeUsd,
  hasNonZeroUsdPriceChanges,
  parseCKPriceDrops,
  parseCKPriceIncreases,
} from "./ckPriceChanges.js";

test("formatPriceChangeUsd formats signed dollar amounts", () => {
  assert.equal(formatPriceChangeUsd(0.5), "+$0.50");
  assert.equal(formatPriceChangeUsd(10), "+$10.00");
  assert.equal(formatPriceChangeUsd(-0.25), "-$0.25");
  assert.equal(formatPriceChangeUsd(-10), "-$10.00");
  assert.equal(formatPriceChangeUsd(0), null);
  assert.equal(formatPriceChangeUsd(null), null);
  assert.equal(formatPriceChangeUsd(Number.NaN), null);
});

test("parseCKPriceIncreases returns top card names with USD changes when present", () => {
  const payload = {
    top: [
      { cardName: "Lightning Bolt", priceChangeUsd: 0.5 },
      { cardName: " Counterspell ", priceChangeUsd: 0.1 },
      { cardName: "", priceChangeUsd: 0.05 },
      { cardName: "Sol Ring", priceChangePercent: 5 },
    ],
  };

  const increases = parseCKPriceIncreases(payload);

  assert.equal(increases.length, 3);
  assert.equal(increases[0].cardName, "Lightning Bolt");
  assert.equal(increases[0].priceChangeUsd, 0.5);
  assert.equal(increases[1].cardName, "Counterspell");
  assert.equal(increases[1].priceChangeUsd, 0.1);
  assert.equal(increases[2].cardName, "Sol Ring");
  assert.equal(increases[2].priceChangeUsd, null);
});

test("parseCKPriceIncreases returns an empty list for invalid payloads", () => {
  assert.deepEqual(parseCKPriceIncreases(null), []);
  assert.deepEqual(parseCKPriceIncreases({}), []);
  assert.deepEqual(parseCKPriceIncreases({ top: "invalid" }), []);
});

test("parseCKPriceDrops returns bottom card names with USD changes when present", () => {
  const payload = {
    bottom: [
      { cardName: "Counterspell", priceChangeUsd: -1.5 },
      { cardName: " Sol Ring ", priceChangeUsd: -0.1 },
      { cardName: "", priceChangeUsd: -0.05 },
      { cardName: "Lightning Bolt", priceChangePercent: -5 },
    ],
  };

  const drops = parseCKPriceDrops(payload);

  assert.equal(drops.length, 3);
  assert.equal(drops[0].cardName, "Counterspell");
  assert.equal(drops[0].priceChangeUsd, -1.5);
  assert.equal(drops[1].cardName, "Sol Ring");
  assert.equal(drops[1].priceChangeUsd, -0.1);
  assert.equal(drops[2].cardName, "Lightning Bolt");
  assert.equal(drops[2].priceChangeUsd, null);
});

test("parseCKPriceDrops returns an empty list for invalid payloads", () => {
  assert.deepEqual(parseCKPriceDrops(null), []);
  assert.deepEqual(parseCKPriceDrops({}), []);
  assert.deepEqual(parseCKPriceDrops({ bottom: "invalid" }), []);
});

test("hasNonZeroUsdPriceChanges is false when all changes are zero, missing, or absent", () => {
  assert.equal(hasNonZeroUsdPriceChanges([], []), false);
  assert.equal(
    hasNonZeroUsdPriceChanges(
      [{ cardName: "Bolt", priceChangeUsd: 0 }],
      [{ cardName: "Ring", priceChangeUsd: 0 }],
    ),
    false,
  );
  assert.equal(
    hasNonZeroUsdPriceChanges(
      [{ cardName: "Bolt", priceChangeUsd: null }],
      [{ cardName: "Ring", priceChangeUsd: null }],
    ),
    false,
  );
});

test("hasNonZeroUsdPriceChanges is true when any listing has a non-zero USD change", () => {
  assert.equal(
    hasNonZeroUsdPriceChanges(
      [{ cardName: "Bolt", priceChangeUsd: 0.5 }],
      [{ cardName: "Ring", priceChangeUsd: 0 }],
    ),
    true,
  );
  assert.equal(
    hasNonZeroUsdPriceChanges(
      [{ cardName: "Bolt", priceChangeUsd: 0 }],
      [{ cardName: "Ring", priceChangeUsd: -0.25 }],
    ),
    true,
  );
});
