import assert from "node:assert/strict";
import { savedCardMatchesFilter } from "./savedCartFilter.js";

const baseCard = {
  name: "Lightning Bolt",
  src: "The Cardboard Collective",
  price: 1.5,
  quality: "NM",
  isFoil: false,
  extraInfo: "Modern Masters",
};

assert.equal(savedCardMatchesFilter(baseCard, ""), true);
assert.equal(savedCardMatchesFilter(baseCard, "   "), true);

assert.equal(savedCardMatchesFilter(baseCard, "lightning"), true);
assert.equal(savedCardMatchesFilter(baseCard, "cardboard"), true);
assert.equal(savedCardMatchesFilter(baseCard, "nm"), true);
assert.equal(savedCardMatchesFilter(baseCard, "modern"), true);

assert.equal(
  savedCardMatchesFilter({ ...baseCard, isFoil: true }, "foil"),
  true,
);
assert.equal(savedCardMatchesFilter(baseCard, "foil"), false);
assert.equal(savedCardMatchesFilter(baseCard, "counterspell"), false);

console.log("savedCartFilter tests passed");
