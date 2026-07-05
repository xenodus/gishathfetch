import assert from "node:assert/strict";
import { LGS_OPTIONS } from "../constants.js";
import {
  storeSelectionsMatch,
  validateFavouriteStores,
} from "./favouriteStores.js";

assert.deepEqual(validateFavouriteStores(["Hideout", "Unknown Store"]), [
  "Hideout",
]);
assert.deepEqual(validateFavouriteStores(null), []);
assert.deepEqual(validateFavouriteStores(LGS_OPTIONS), LGS_OPTIONS);

assert.equal(storeSelectionsMatch(["Hideout"], ["Hideout"]), true);
assert.equal(
  storeSelectionsMatch(["Hideout", "OneMtg"], ["OneMtg", "Hideout"]),
  true,
);
assert.equal(storeSelectionsMatch(["Hideout"], ["Hideout", "OneMtg"]), false);
assert.equal(storeSelectionsMatch([], []), true);

console.log("favouriteStores tests passed");
