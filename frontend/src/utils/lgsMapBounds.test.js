import assert from "node:assert/strict";
import test from "node:test";
import { computeLgsMapBounds } from "./lgsMapBounds.js";

test("computeLgsMapBounds fits all marker coordinates", () => {
  const stores = [
    { lat: 1.3, lng: 103.8 },
    { lat: 1.35, lng: 103.9 },
  ];

  assert.deepEqual(computeLgsMapBounds(stores), {
    west: 103.8,
    east: 103.9,
    north: 1.35,
    south: 1.3,
  });
});
