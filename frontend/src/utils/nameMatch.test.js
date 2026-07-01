import assert from "node:assert/strict";
import { getNameMatchTier, NAME_MATCH_TIER } from "./nameMatch.js";

assert.equal(
  getNameMatchTier("Cosmic Cube", "Cosmic Cube"),
  NAME_MATCH_TIER.EXACT,
);
assert.equal(
  getNameMatchTier("Cosmic Cube", "cosmic cube"),
  NAME_MATCH_TIER.EXACT,
);
assert.equal(
  getNameMatchTier("Cosmic Cube Foil", "Cosmic Cube"),
  NAME_MATCH_TIER.PREFIX,
);
assert.equal(
  getNameMatchTier("Construct a Cosmic Cube", "Cosmic Cube"),
  NAME_MATCH_TIER.PARTIAL,
);
assert.equal(
  getNameMatchTier("Lightning Bolt", "Cosmic Cube"),
  NAME_MATCH_TIER.NONE,
);
assert.equal(getNameMatchTier("Opt", ""), NAME_MATCH_TIER.NONE);

console.log("nameMatch tests passed");
