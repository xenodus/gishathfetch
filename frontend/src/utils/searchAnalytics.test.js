import assert from "node:assert/strict";
import {
  cardNamesMatchForVerify,
  foldCardNameForMatch,
  listStoresWithNoResults,
} from "./searchAnalytics.js";

assert.equal(
  foldCardNameForMatch("Kíli the Resourceful"),
  foldCardNameForMatch("Kili the Resourceful"),
);
assert.equal(foldCardNameForMatch("Juzám Djinn"), "juzam djinn");

assert.equal(cardNamesMatchForVerify("Lightning Bolt", "lightning bolt"), true);
assert.equal(
  cardNamesMatchForVerify("Kíli the Resourceful", "Kili the Resourceful"),
  true,
);
assert.equal(cardNamesMatchForVerify("Lightning Bolt", "Counterspell"), false);

assert.deepEqual(
  listStoresWithNoResults(
    ["Alpha", "Beta"],
    [
      { store: "Alpha", itemCount: 0 },
      { store: "Beta", itemCount: 2 },
    ],
    [],
  ),
  ["Alpha"],
);

assert.deepEqual(
  listStoresWithNoResults(
    ["Alpha", "Beta"],
    [
      { store: "Alpha", itemCount: 0 },
      { store: "Beta", itemCount: 0 },
    ],
    [{ store: "Beta", error: "timeout" }],
  ),
  ["Alpha"],
);

assert.deepEqual(
  listStoresWithNoResults(["Alpha"], [{ store: "Alpha", itemCount: 1 }], []),
  [],
);

console.log("searchAnalytics tests passed");
