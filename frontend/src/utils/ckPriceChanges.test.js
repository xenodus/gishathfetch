import assert from "node:assert/strict";
import test from "node:test";
import {
  formatPriceChangePercent,
  parseCKPriceDrops,
  parseCKPriceIncreases,
} from "./ckPriceChanges.js";

test("formatPriceChangePercent formats whole and fractional percentages", () => {
  assert.equal(formatPriceChangePercent(15), "15%");
  assert.equal(formatPriceChangePercent(12.34), "12.3%");
  assert.equal(formatPriceChangePercent(-10), "-10%");
  assert.equal(formatPriceChangePercent(-10, { absolute: true }), "10%");
  assert.equal(formatPriceChangePercent(null), null);
  assert.equal(formatPriceChangePercent(Number.NaN), null);
});

test("parseCKPriceIncreases returns top card names up to the display limit", () => {
  const payload = {
    top: [
      { cardName: "Lightning Bolt", priceChangePercent: 15 },
      { cardName: " Counterspell ", priceChangePercent: 10 },
      { cardName: "", priceChangePercent: 5 },
      { cardName: "Sol Ring", priceChangePercent: 5 },
    ],
  };

  const increases = parseCKPriceIncreases(payload);

  assert.equal(increases.length, 3);
  assert.equal(increases[0].cardName, "Lightning Bolt");
  assert.equal(increases[0].priceChangePercent, 15);
  assert.equal(increases[1].cardName, "Counterspell");
  assert.equal(increases[1].priceChangePercent, 10);
});

test("parseCKPriceIncreases returns an empty list for invalid payloads", () => {
  assert.deepEqual(parseCKPriceIncreases(null), []);
  assert.deepEqual(parseCKPriceIncreases({}), []);
  assert.deepEqual(parseCKPriceIncreases({ top: "invalid" }), []);
});

test("parseCKPriceDrops returns bottom card names up to the display limit", () => {
  const payload = {
    bottom: [
      { cardName: "Counterspell", priceChangePercent: -15 },
      { cardName: " Sol Ring ", priceChangePercent: -10 },
      { cardName: "", priceChangePercent: -5 },
      { cardName: "Lightning Bolt", priceChangePercent: -5 },
    ],
  };

  const drops = parseCKPriceDrops(payload);

  assert.equal(drops.length, 3);
  assert.equal(drops[0].cardName, "Counterspell");
  assert.equal(drops[0].priceChangePercent, -15);
  assert.equal(drops[1].cardName, "Sol Ring");
  assert.equal(drops[1].priceChangePercent, -10);
});

test("parseCKPriceDrops returns an empty list for invalid payloads", () => {
  assert.deepEqual(parseCKPriceDrops(null), []);
  assert.deepEqual(parseCKPriceDrops({}), []);
  assert.deepEqual(parseCKPriceDrops({ bottom: "invalid" }), []);
});
