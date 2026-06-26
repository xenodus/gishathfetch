import { getCachedJson, isCacheFresh, setCachedJson } from "./marketCache";

const CK_PRICELIST = "https://api.cardkingdom.com/api/v2/pricelist";
const SCRYFALL_AUTOCOMPLETE = "https://api.scryfall.com/cards/autocomplete";
const SCRYFALL_NAMED = "https://api.scryfall.com/cards/named";

const CK_CACHE_KEY = "ck-pricelist";
const CK_CACHE_TTL = 60 * 60 * 1000;

const buildCheapestByName = (products) => {
  const cheapestByName = new Map();

  for (const product of products) {
    const nameKey = product.name?.trim().toLowerCase();
    if (!nameKey) {
      continue;
    }

    const priceUsd = Number.parseFloat(product.price_retail);
    if (!Number.isFinite(priceUsd) || priceUsd <= 0) {
      continue;
    }

    const existing = cheapestByName.get(nameKey);
    if (!existing || priceUsd < existing.priceUsd) {
      cheapestByName.set(nameKey, {
        cardName: product.name,
        edition: product.edition,
        priceUsd,
        url: `https://www.cardkingdom.com/${product.url}`,
        quantity: Number.parseInt(product.qty_retail, 10),
        isFoil: product.is_foil === "true",
      });
    }
  }

  return cheapestByName;
};

const fetchCkPricelist = async (signal) => {
  const response = await fetch(CK_PRICELIST, { signal });
  if (!response.ok) {
    throw new Error("Card Kingdom price list is unavailable.");
  }

  const payload = await response.json();
  const products = Array.isArray(payload) ? payload : (payload.data ?? []);
  const cheapestByName = buildCheapestByName(products);

  await setCachedJson(
    CK_CACHE_KEY,
    { cheapestByName: [...cheapestByName] },
    CK_CACHE_TTL,
  );

  return cheapestByName;
};

const getCheapestByName = async (signal) => {
  const cached = await getCachedJson(CK_CACHE_KEY);
  if (isCacheFresh(cached)) {
    return new Map(cached.data.cheapestByName);
  }
  return fetchCkPricelist(signal);
};

export const verifyCardName = async (query, signal) => {
  const trimmed = query.trim();
  if (!trimmed) {
    return null;
  }

  const autocompleteResponse = await fetch(
    `${SCRYFALL_AUTOCOMPLETE}?q=${encodeURIComponent(trimmed)}`,
    { signal },
  );
  if (autocompleteResponse.ok) {
    const autocomplete = await autocompleteResponse.json();
    const exactMatch = autocomplete.data?.find(
      (name) => name.toLowerCase() === trimmed.toLowerCase(),
    );
    if (exactMatch) {
      return exactMatch;
    }
  }

  const namedResponse = await fetch(
    `${SCRYFALL_NAMED}?exact=${encodeURIComponent(trimmed)}`,
    { signal },
  );
  if (!namedResponse.ok) {
    return null;
  }

  const card = await namedResponse.json();
  return card.name ?? null;
};

export const fetchCardKingdomLatestPrice = async (cardName, signal) => {
  const cheapestByName = await getCheapestByName(signal);
  const match = cheapestByName.get(cardName.trim().toLowerCase());
  if (!match) {
    return null;
  }
  return match;
};
