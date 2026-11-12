import { describe, expect, it } from "vitest";
import { savedCardMatchesFilter } from "./savedCartFilter";

describe("savedCardMatchesFilter", () => {
  const baseCard = {
    name: "Lightning Bolt",
    src: "The Cardboard Collective",
    price: 1.5,
    quality: "NM",
    isFoil: false,
    extraInfo: "Modern Masters",
  };

  it("matches all cards when query is empty", () => {
    expect(savedCardMatchesFilter(baseCard, "")).toBe(true);
    expect(savedCardMatchesFilter(baseCard, "   ")).toBe(true);
  });

  it("matches card name, store, condition, and set info", () => {
    expect(savedCardMatchesFilter(baseCard, "lightning")).toBe(true);
    expect(savedCardMatchesFilter(baseCard, "cardboard")).toBe(true);
    expect(savedCardMatchesFilter(baseCard, "nm")).toBe(true);
    expect(savedCardMatchesFilter(baseCard, "modern")).toBe(true);
  });

  it("matches foil cards when searching foil", () => {
    expect(savedCardMatchesFilter({ ...baseCard, isFoil: true }, "foil")).toBe(
      true,
    );
    expect(savedCardMatchesFilter(baseCard, "foil")).toBe(false);
  });

  it("returns false when nothing matches", () => {
    expect(savedCardMatchesFilter(baseCard, "counterspell")).toBe(false);
  });
});
