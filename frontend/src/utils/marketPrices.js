import { getCachedJson, isCacheFresh, setCachedJson } from "./marketCache";

const CK_PRICELIST = "https://api.cardkingdom.com/api/v2/pricelist";
const FX_URL = "https://api.frankfurter.app/latest?from=USD&to=SGD";

const CK_CACHE_KEY = "ck-pricelist";
const CK_CACHE_TTL = 6 * 60 * 60 * 1000;
const FX_CACHE_KEY = "usd-sgd-rate";
const FX_CACHE_TTL = 12 * 60 * 60 * 1000;
const JOURNAL_TTL = 365 * 24 * 60 * 60 * 1000;

const normalizeText = (value) =>
  (value ?? "")
    .toLowerCase()
    .replace(/[[\]()]/g, " ")
    .replace(/\s+/g, " ")
    .trim();

const productKey = (scryfallId, isFoil) => `${scryfallId}:${isFoil}`;

const journalCacheKey = (key) => `ck-journal-${key}`;

const scoreProductMatch = (product, { name, extraInfo, isFoil }) => {
  let score = 0;
  if (product.name?.toLowerCase() === name.toLowerCase()) {
    score += 5;
  }
  const productFoil = product.is_foil === "true";
  if (productFoil === isFoil) {
    score += 4;
  }
  const extra = normalizeText(extraInfo);
  const edition = normalizeText(product.edition);
  const variation = normalizeText(product.variation);
  if (extra && (extra.includes(edition) || extra.includes(variation))) {
    score += 6;
  }
  return score;
};

const findCardKingdomProduct = (index, card) => {
  const candidates = [];
  for (const [key, product] of index.entries()) {
    if (!key.endsWith(`:${card.isFoil}`)) {
      continue;
    }
    const score = scoreProductMatch(product, card);
    if (score > 0) {
      candidates.push({ product, score });
    }
  }
  if (candidates.length === 0) {
    return null;
  }
  candidates.sort((a, b) => b.score - a.score);
  return candidates[0].product;
};

const fetchCkPricelist = async (onProgress) => {
  onProgress?.("Downloading Card Kingdom price list…");
  const response = await fetch(CK_PRICELIST);
  if (!response.ok) {
    throw new Error("Card Kingdom price list is unavailable.");
  }
  const payload = await response.json();
  const products = Array.isArray(payload) ? payload : (payload.data ?? []);
  const meta = payload.meta ?? {};
  const byLookupKey = new Map();

  for (const product of products) {
    if (!product.scryfall_id) {
      continue;
    }
    const key = productKey(product.scryfall_id, product.is_foil === "true");
    if (!byLookupKey.has(key)) {
      byLookupKey.set(key, {
        scryfall_id: product.scryfall_id,
        is_foil: product.is_foil === "true",
        name: product.name,
        edition: product.edition,
        variation: product.variation,
        price_retail: product.price_retail,
        qty_retail: product.qty_retail,
        url: product.url,
      });
    }
  }

  await setCachedJson(
    CK_CACHE_KEY,
    {
      byLookupKey: [...byLookupKey],
      priceListDate: meta.created_at ?? new Date().toISOString().slice(0, 10),
    },
    CK_CACHE_TTL,
  );
  return {
    index: byLookupKey,
    priceListDate: meta.created_at ?? new Date().toISOString().slice(0, 10),
  };
};

const getCkIndex = async (onProgress) => {
  const cached = await getCachedJson(CK_CACHE_KEY);
  if (isCacheFresh(cached)) {
    return {
      index: new Map(cached.data.byLookupKey),
      priceListDate:
        cached.data.priceListDate ?? new Date().toISOString().slice(0, 10),
    };
  }
  return fetchCkPricelist(onProgress);
};

const toSnapshotDate = (priceListDate) => {
  if (!priceListDate) {
    return new Date().toISOString().slice(0, 10);
  }
  return priceListDate.slice(0, 10);
};

const getJournalPoints = async (key) => {
  const cached = await getCachedJson(journalCacheKey(key));
  return cached?.data?.points ?? [];
};

const recordJournalSnapshot = async (key, price, date) => {
  if (!Number.isFinite(price)) {
    return getJournalPoints(key);
  }
  const points = await getJournalPoints(key);
  const nextPoints = points.filter((point) => point.date !== date);
  nextPoints.push({ date, price });
  nextPoints.sort((a, b) => a.date.localeCompare(b.date));
  await setCachedJson(
    journalCacheKey(key),
    { points: nextPoints.slice(-120) },
    JOURNAL_TTL,
  );
  return nextPoints;
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

export const loadMarketSnapshot = async (card, onProgress) => {
  onProgress?.("Looking up Card Kingdom price…");
  const [{ index, priceListDate }, usdToSgd] = await Promise.all([
    getCkIndex(onProgress),
    getUsdToSgdRate(),
  ]);

  const product = findCardKingdomProduct(index, card);
  if (!product) {
    throw new Error("This printing is not listed on Card Kingdom.");
  }

  const key = productKey(product.scryfall_id, product.is_foil);
  const priceUsd = Number.parseFloat(product.price_retail);
  const snapshotDate = toSnapshotDate(priceListDate);
  const history = await recordJournalSnapshot(key, priceUsd, snapshotDate);

  return {
    cardName: product.name,
    setName: product.edition,
    collectorNumber: null,
    image: card.img,
    finish: product.is_foil ? "foil" : "normal",
    isFoil: product.is_foil,
    usdToSgd,
    priceListDate: snapshotDate,
    references: {
      cardkingdom: {
        usd: priceUsd,
        sgd: priceUsd * usdToSgd,
        url: `https://www.cardkingdom.com/${product.url}`,
        quantity: Number.parseInt(product.qty_retail, 10),
        source: "live",
      },
    },
    history: {
      cardkingdom: history,
    },
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
