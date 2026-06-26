import { getCachedJson, isCacheFresh, setCachedJson } from "./marketCache";

const SCRYFALL_SEARCH = "https://api.scryfall.com/cards/search";
const MTGJSON_SET = "https://mtgjson.com/api/v5";
const MTGJSON_ALL_PRICES_GZ = `${MTGJSON_SET}/AllPrices.json.gz`;
const CK_PRICELIST = "https://api.cardkingdom.com/api/v2/pricelist";
const FX_URL = "https://api.frankfurter.app/latest?from=USD&to=SGD";

const CK_CACHE_KEY = "ck-pricelist";
const CK_CACHE_TTL = 6 * 60 * 60 * 1000;
const HISTORY_CACHE_KEY = "mtgjson-allprices-gz";
const HISTORY_CACHE_TTL = 24 * 60 * 60 * 1000;
const FX_CACHE_KEY = "usd-sgd-rate";
const FX_CACHE_TTL = 12 * 60 * 60 * 1000;

const finishKey = (isFoil) => (isFoil ? "foil" : "normal");

const normalizeText = (value) =>
  (value ?? "")
    .toLowerCase()
    .replace(/[[\]()]/g, " ")
    .replace(/\s+/g, " ")
    .trim();

const scoreScryfallMatch = (card, { name, extraInfo, isFoil }) => {
  let score = 0;
  if (card.name?.toLowerCase() === name.toLowerCase()) {
    score += 5;
  }
  const cardFoil = Boolean(card.finishes?.includes("foil"));
  if (cardFoil === isFoil) {
    score += 4;
  }
  const extra = normalizeText(extraInfo);
  const setName = normalizeText(card.set_name);
  const setCode = normalizeText(card.set);
  if (extra && (extra.includes(setName) || extra.includes(setCode))) {
    score += 6;
  }
  return score;
};

export const findScryfallPrinting = async ({ name, extraInfo, isFoil }) => {
  const query = encodeURIComponent(`!"${name}"`);
  const response = await fetch(
    `${SCRYFALL_SEARCH}?q=${query}&unique=prints&order=released`,
  );
  if (!response.ok) {
    throw new Error("Could not find this card on Scryfall.");
  }
  const payload = await response.json();
  const cards = payload.data ?? [];
  if (cards.length === 0) {
    throw new Error("No Scryfall printings found for this card.");
  }
  const ranked = [...cards].sort(
    (a, b) =>
      scoreScryfallMatch(b, { name, extraInfo, isFoil }) -
      scoreScryfallMatch(a, { name, extraInfo, isFoil }),
  );
  return ranked[0];
};

export const resolveMtgjsonUuid = async (scryfallCard) => {
  const setCode = scryfallCard.set?.toUpperCase();
  if (!setCode) {
    throw new Error("Missing set code for market lookup.");
  }
  const response = await fetch(`${MTGJSON_SET}/${setCode}.json`);
  if (!response.ok) {
    throw new Error("Could not load MTGJSON set data.");
  }
  const payload = await response.json();
  const match = (payload.data?.cards ?? []).find(
    (card) => card.identifiers?.scryfallId === scryfallCard.id,
  );
  if (!match?.uuid) {
    throw new Error("Could not map this printing to MTGJSON price data.");
  }
  return match.uuid;
};

const fetchCkPricelist = async (onProgress) => {
  onProgress?.("Downloading Card Kingdom price list…");
  const response = await fetch(CK_PRICELIST);
  if (!response.ok) {
    throw new Error("Card Kingdom price list is unavailable.");
  }
  const products = await response.json();
  const byScryfallId = new Map();
  for (const product of products) {
    if (!product.scryfall_id) {
      continue;
    }
    const key = `${product.scryfall_id}:${product.is_foil === "true"}`;
    if (!byScryfallId.has(key)) {
      byScryfallId.set(key, {
        price_retail: product.price_retail,
        qty_retail: product.qty_retail,
        url: product.url,
      });
    }
  }
  await setCachedJson(
    CK_CACHE_KEY,
    { byScryfallId: [...byScryfallId] },
    CK_CACHE_TTL,
  );
  return byScryfallId;
};

const getCkIndex = async (onProgress) => {
  const cached = await getCachedJson(CK_CACHE_KEY);
  if (isCacheFresh(cached)) {
    return new Map(cached.data.byScryfallId);
  }
  return fetchCkPricelist(onProgress);
};

export const lookupCardKingdomPrice = async (
  scryfallId,
  isFoil,
  onProgress,
) => {
  const index = await getCkIndex(onProgress);
  const product = index.get(`${scryfallId}:${isFoil}`);
  if (!product) {
    return null;
  }
  return {
    priceUsd: Number.parseFloat(product.price_retail),
    quantity: Number.parseInt(product.qty_retail, 10),
    url: `https://www.cardkingdom.com/${product.url}`,
    updatedAt: null,
  };
};

const fetchHistoryBlob = async (onProgress) => {
  onProgress?.(
    "Downloading market price history (first load may take a minute)…",
  );
  const response = await fetch(MTGJSON_ALL_PRICES_GZ);
  if (!response.ok) {
    throw new Error("MTGJSON price history is unavailable.");
  }
  const blob = await response.blob();
  await setCachedJson(HISTORY_CACHE_KEY, { blob: blob }, HISTORY_CACHE_TTL);
  return blob;
};

const getHistoryBlob = async (onProgress) => {
  const cached = await getCachedJson(HISTORY_CACHE_KEY);
  if (isCacheFresh(cached)) {
    return cached.data.blob;
  }
  return fetchHistoryBlob(onProgress);
};

const extractUuidHistory = async (blob, uuid) => {
  const stream = blob.stream().pipeThrough(new DecompressionStream("gzip"));
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  const needle = `"${uuid}":`;
  let found = false;

  while (true) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
    if (!found) {
      const idx = buffer.indexOf(needle);
      if (idx === -1) {
        buffer = buffer.slice(-(needle.length + 10));
        continue;
      }
      found = true;
      buffer = buffer.slice(idx + needle.length).trimStart();
    }
    if (found) {
      if (!buffer.startsWith("{")) {
        throw new Error("Unexpected MTGJSON payload while reading history.");
      }
      let depth = 0;
      let end = 0;
      for (let i = 0; i < buffer.length; i += 1) {
        const ch = buffer[i];
        if (ch === "{") {
          depth += 1;
        } else if (ch === "}") {
          depth -= 1;
          if (depth === 0) {
            end = i + 1;
            break;
          }
        }
      }
      if (end > 0) {
        return JSON.parse(buffer.slice(0, end));
      }
    }
  }
  throw new Error("No price history found for this printing.");
};

const seriesFromProvider = (providerData, finish) => {
  const retail = providerData?.retail?.[finish] ?? {};
  return Object.entries(retail)
    .map(([date, price]) => ({ date, price: Number(price) }))
    .filter((point) => Number.isFinite(point.price))
    .sort((a, b) => a.date.localeCompare(b.date));
};

export const fetchPriceHistory = async (uuid, isFoil, onProgress) => {
  const blob = await getHistoryBlob(onProgress);
  onProgress?.("Reading price history for this printing…");
  const payload = await extractUuidHistory(blob, uuid);
  const finish = finishKey(isFoil);
  return {
    cardkingdom: seriesFromProvider(payload.paper?.cardkingdom, finish),
    tcgplayer: seriesFromProvider(payload.paper?.tcgplayer, finish),
  };
};

export const getUsdToSgdRate = async () => {
  const cached = await getCachedJson(FX_CACHE_KEY);
  if (isCacheFresh(cached)) {
    return cached.data.rate;
  }
  const response = await fetch(FX_URL);
  if (!response.ok) {
    return 1.35;
  }
  const payload = await response.json();
  const rate = payload?.rates?.SGD;
  const resolved = Number.isFinite(rate) ? rate : 1.35;
  await setCachedJson(FX_CACHE_KEY, { rate: resolved }, FX_CACHE_TTL);
  return resolved;
};

const scryfallSpotPrice = (scryfallCard, isFoil) => {
  const prices = scryfallCard.prices ?? {};
  const raw = isFoil ? (prices.usd_foil ?? prices.usd) : prices.usd;
  const value = Number.parseFloat(raw);
  return Number.isFinite(value) ? value : null;
};

export const loadMarketSnapshot = async (card, onProgress) => {
  onProgress?.("Looking up card printing…");
  const scryfallCard = await findScryfallPrinting(card);
  const uuid = await resolveMtgjsonUuid(scryfallCard);
  const [cardKingdom, history, usdToSgd] = await Promise.all([
    lookupCardKingdomPrice(scryfallCard.id, card.isFoil, onProgress),
    fetchPriceHistory(uuid, card.isFoil, onProgress),
    getUsdToSgdRate(),
  ]);
  const tcgplayerSpot = scryfallSpotPrice(scryfallCard, card.isFoil);
  const finish = finishKey(card.isFoil);
  const tcgHistory = history.tcgplayer;
  const ckHistory = history.cardkingdom;
  const latestTcg =
    tcgHistory.length > 0
      ? tcgHistory[tcgHistory.length - 1].price
      : tcgplayerSpot;
  const latestCk =
    cardKingdom?.priceUsd ??
    (ckHistory.length > 0 ? ckHistory[ckHistory.length - 1].price : null);

  return {
    cardName: scryfallCard.name,
    setName: scryfallCard.set_name,
    collectorNumber: scryfallCard.collector_number,
    image: scryfallCard.image_uris?.normal ?? card.img,
    finish,
    isFoil: card.isFoil,
    usdToSgd,
    references: {
      cardkingdom: {
        usd: latestCk,
        sgd: latestCk == null ? null : latestCk * usdToSgd,
        url: cardKingdom?.url ?? null,
        quantity: cardKingdom?.quantity ?? null,
        source: cardKingdom ? "live" : "history",
      },
      tcgplayer: {
        usd: latestTcg,
        sgd: latestTcg == null ? null : latestTcg * usdToSgd,
        url: scryfallCard.purchase_uris?.tcgplayer ?? null,
        lowUsd: Number.parseFloat(scryfallCard.prices?.usd) || null,
        highUsd: null,
        source: "history",
      },
    },
    history,
    scryfallId: scryfallCard.id,
    tcgplayerId: scryfallCard.tcgplayer_id ?? null,
  };
};

export const filterHistoryByRange = (series, range) => {
  if (!series?.length || range === "all") {
    return series ?? [];
  }
  const days = range === "1w" ? 7 : range === "1m" ? 30 : 90;
  const latest = new Date(`${series[series.length - 1].date}T00:00:00Z`);
  const cutoff = new Date(latest);
  cutoff.setUTCDate(cutoff.getUTCDate() - days);
  return series.filter(
    (point) => new Date(`${point.date}T00:00:00Z`) >= cutoff,
  );
};

export const computeTrendPercent = (series) => {
  if (!series || series.length < 2) {
    return null;
  }
  const first = series[0].price;
  const last = series[series.length - 1].price;
  if (!first) {
    return null;
  }
  return ((last - first) / first) * 100;
};
