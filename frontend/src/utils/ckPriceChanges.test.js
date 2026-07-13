import assert from "node:assert/strict";
import test from "node:test";
import { parseCKPriceIncreases } from "./ckPriceChanges.js";

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
