import type { SearchResponse } from "../api/types";

/** Placeholder results for UI development before live API integration. */
export function mockSearchResponse(query: string): SearchResponse {
  const normalized = query.trim();
  return {
    data: [
      {
        name: normalized,
        url: "https://5-mana.sg/",
        img: "https://cards.scryfall.io/normal/front/0/0/Opt.jpg",
        price: 1.5,
        inStock: true,
        isFoil: false,
        src: "5 Mana",
        quality: "NM",
      },
      {
        name: normalized,
        url: "https://hideoutcg.com/",
        img: "https://cards.scryfall.io/normal/front/0/0/Opt.jpg",
        price: 2.0,
        inStock: true,
        isFoil: false,
        src: "Hideout",
        quality: "NM",
      },
      {
        name: normalized,
        url: "https://www.moxandlotus.sg/",
        img: "https://cards.scryfall.io/normal/front/0/0/Opt.jpg",
        price: 3.25,
        inStock: true,
        isFoil: true,
        src: "Mox & Lotus",
        quality: "NM",
        extraInfo: "Foil",
      },
    ],
    errors: [
      {
        store: "Example Store",
        error: "Mock error row for UI testing",
        statusCode: 503,
      },
    ],
    stats: [
      { store: "5 Mana", itemCount: 1, durationMs: 420 },
      { store: "Hideout", itemCount: 1, durationMs: 890 },
    ],
    totalDurationMs: 1200,
    cardKingdomPrice: {
      name: normalized,
      price: 1.99,
      url: "https://www.cardkingdom.com/",
    },
  };
}
