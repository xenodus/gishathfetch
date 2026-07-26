import assert from "node:assert/strict";
import { cardIdentityKey } from "./cardIdentity.js";
import {
  CART_TRANSFER_PREFIX,
  decodeCartImport,
  encodeCartExport,
  mergeCartImport,
  validateCartItems,
} from "./cartTransfer.js";

const sampleCard = {
  name: "Lightning Bolt",
  src: "Hideout",
  url: "https://example.com/bolt",
  price: 12.5,
  quality: "NM",
  isFoil: false,
  inStock: true,
  savedAt: 1000,
};

assert.deepEqual(validateCartItems([sampleCard, { name: "x" }]), [sampleCard]);

const encoded = encodeCartExport([sampleCard]);
assert.ok(encoded.startsWith(CART_TRANSFER_PREFIX));

const decoded = decodeCartImport(encoded);
assert.equal(decoded.error, null);
assert.equal(decoded.items.length, 1);
assert.equal(decoded.items[0].name, "Lightning Bolt");

const roundTrip = decodeCartImport(encoded);
assert.equal(cardIdentityKey(roundTrip.items[0]), cardIdentityKey(sampleCard));

const merged = mergeCartImport(
  [sampleCard],
  [{ ...sampleCard, price: 9, savedAt: 2000 }],
);
assert.equal(merged.length, 1);
assert.equal(merged[0].price, 9);

const appended = mergeCartImport(
  [sampleCard],
  [
    {
      ...sampleCard,
      name: "Counterspell",
      url: "https://example.com/counter",
    },
  ],
);
assert.equal(appended.length, 2);

assert.match(decodeCartImport("bad").error, /Unrecognized/);
assert.match(decodeCartImport("").error, /Paste/);

console.log("cartTransfer tests passed");
