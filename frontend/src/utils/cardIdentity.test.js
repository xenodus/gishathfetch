import assert from "node:assert/strict";
import {
  cardIdentityKey,
  cardsExactMatch,
  dedupeCartItems,
  normalizeExtraInfo,
} from "./cardIdentity.js";

const base = {
  src: "The TCG Marketplace",
  name: "Swordsman's Steel",
  extraInfo: "[Marvel Super Heroes Commander]",
};

assert.equal(
  normalizeExtraInfo(["[Marvel Super Heroes Commander]"]),
  "[Marvel Super Heroes Commander]",
);
assert.equal(
  normalizeExtraInfo("[Marvel Super Heroes Commander] "),
  "[Marvel Super Heroes Commander]",
);
assert.equal(normalizeExtraInfo([]), "");
assert.equal(normalizeExtraInfo("[]"), "");
assert.equal(normalizeExtraInfo("()"), "");
assert.equal(normalizeExtraInfo("[ ]"), "");
assert.equal(normalizeExtraInfo(undefined), "");
assert.equal(normalizeExtraInfo("[Set B] [Set A]"), "[Set A] [Set B]");

assert.equal(
  normalizeExtraInfo("[Modern Horizons 2] (Borderless)"),
  "(Borderless) [Modern Horizons 2]",
);
assert.notEqual(
  normalizeExtraInfo("[Modern Horizons 2] (Borderless)"),
  normalizeExtraInfo("[Modern Horizons 2] (Retro Frame)"),
);
assert.notEqual(
  normalizeExtraInfo("[MH2] Borderless"),
  normalizeExtraInfo("[MH2] Retro Frame"),
);

assert.ok(
  cardsExactMatch(base, {
    ...base,
    extraInfo: ["[Marvel Super Heroes Commander]"],
  }),
);
assert.ok(
  cardsExactMatch(base, {
    ...base,
    name: "Swordsman\u2019s Steel",
  }),
);
assert.ok(
  cardsExactMatch(base, {
    ...base,
    url: "https://example.com/a",
    price: 20,
  }),
);
assert.ok(
  !cardsExactMatch(base, {
    ...base,
    isFoil: true,
  }),
);
assert.ok(
  !cardsExactMatch(base, {
    ...base,
    quality: "NM",
  }),
);

const deduped = dedupeCartItems([
  { ...base, savedAt: 2, url: "https://example.com/new" },
  { ...base, savedAt: 1, url: "https://example.com/old" },
  {
    ...base,
    name: "Other Card",
    savedAt: 3,
  },
]);
assert.equal(deduped.length, 2);
assert.equal(deduped[0].url, "https://example.com/new");

const legacyOrder = dedupeCartItems([
  { ...base, savedAt: 1, url: "https://example.com/old" },
  { ...base, savedAt: 2, url: "https://example.com/new" },
]);
assert.equal(legacyOrder.length, 1);
assert.equal(legacyOrder[0].url, "https://example.com/new");

assert.equal(cardIdentityKey(base), cardIdentityKey({ ...base, savedAt: 99 }));

console.log("cardIdentity tests passed");
