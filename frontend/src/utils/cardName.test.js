import assert from "node:assert/strict";
import test from "node:test";
import {
  applyCanonicalCardNames,
  cardNamesMatchIgnoringDiacritics,
  foldCardNameForMatch,
} from "./cardName.js";

test("foldCardNameForMatch strips diacritics for comparison", () => {
  assert.equal(
    foldCardNameForMatch("Kíli the Resourceful"),
    foldCardNameForMatch("Kili the Resourceful"),
  );
  assert.equal(foldCardNameForMatch("Juzám Djinn"), "juzam djinn");
  assert.equal(
    foldCardNameForMatch("Lim-Dûl the Necromancer"),
    "lim-dul the necromancer",
  );
  assert.equal(
    foldCardNameForMatch("Palantír of Orthanc"),
    "palantir of orthanc",
  );
  assert.equal(foldCardNameForMatch("Ifh-Bíff Efreet"), "ifh-biff efreet");
  assert.equal(
    foldCardNameForMatch("Círdan the Shipwright"),
    "cirdan the shipwright",
  );
  assert.equal(foldCardNameForMatch("Gríma Wormtongue"), "grima wormtongue");
  assert.equal(foldCardNameForMatch("Séance"), "seance");
});

test("cardNamesMatchIgnoringDiacritics matches accented and ASCII variants", () => {
  assert.equal(
    cardNamesMatchIgnoringDiacritics(
      "Kíli the Resourceful",
      "Kili the Resourceful",
    ),
    true,
  );
  assert.equal(
    cardNamesMatchIgnoringDiacritics("Juzám Djinn", "Juzam Djinn"),
    true,
  );
  assert.equal(
    cardNamesMatchIgnoringDiacritics("Lightning Bolt", "Counterspell"),
    false,
  );
});

test("applyCanonicalCardNames replaces only changed names", () => {
  const resolved = new Map([["Kili the Resourceful", "Kíli the Resourceful"]]);
  const cards = [
    { name: "Kili the Resourceful", price: 1 },
    { name: "Lightning Bolt", price: 2 },
  ];

  const updated = applyCanonicalCardNames(cards, resolved);
  assert.equal(updated[0].name, "Kíli the Resourceful");
  assert.equal(updated[1].name, "Lightning Bolt");
});
