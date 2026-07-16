export function formatPriceChangeUsd(usd) {
  if (typeof usd !== "number" || !Number.isFinite(usd)) {
    return null;
  }

  const rounded = Math.round(usd * 100) / 100;
  const amount = Math.abs(rounded).toFixed(2);

  if (rounded === 0) {
    return null;
  }
  if (rounded > 0) {
    return `+$${amount}`;
  }
  return `-$${amount}`;
}

function listingDedupeKey(item) {
  const url = typeof item?.url === "string" ? item.url.trim() : "";
  if (url) {
    return url;
  }

  const cardName =
    typeof item?.cardName === "string"
      ? item.cardName.trim().toLowerCase()
      : "";
  const edition =
    typeof item?.edition === "string" ? item.edition.trim().toLowerCase() : "";
  if (cardName) {
    return `${cardName}|${edition}|${item?.isFoil ? "foil" : "nonfoil"}`;
  }

  const nameKey =
    typeof item?.nameKey === "string" ? item.nameKey.trim().toLowerCase() : "";
  return nameKey || null;
}

function dedupePriceChangeListings(listings, limit) {
  const seen = new Set();
  const deduped = [];

  for (const item of listings) {
    const dedupeKey = listingDedupeKey(item);
    if (!dedupeKey || seen.has(dedupeKey)) {
      continue;
    }

    seen.add(dedupeKey);
    deduped.push(item);
    if (deduped.length >= limit) {
      break;
    }
  }

  return deduped;
}

function parseCKPriceChangeListings(listings, matchesDirection, limit = 20) {
  if (!Array.isArray(listings)) {
    return [];
  }

  const parsed = listings
    .map((item, index) => {
      const cardName =
        typeof item?.cardName === "string" ? item.cardName.trim() : "";
      const nameKey =
        typeof item?.nameKey === "string" ? item.nameKey.trim() : "";

      return {
        id: nameKey || (cardName ? `${cardName}-${index}` : `listing-${index}`),
        cardName,
        priceChangeUsd:
          typeof item?.priceChangeUsd === "number" &&
          Number.isFinite(item.priceChangeUsd)
            ? item.priceChangeUsd
            : null,
        source: item,
      };
    })
    .filter(
      (item) =>
        item.cardName.length > 0 &&
        item.priceChangeUsd !== null &&
        matchesDirection(item.priceChangeUsd),
    );

  return dedupePriceChangeListings(parsed, limit).map(
    ({ id, cardName, priceChangeUsd }) => ({
      id,
      cardName,
      priceChangeUsd,
    }),
  );
}

export function parseCKPriceIncreases(payload, limit = 20) {
  return parseCKPriceChangeListings(
    payload?.top,
    (priceChangeUsd) => priceChangeUsd > 0,
    limit,
  );
}

export function parseCKPriceDrops(payload, limit = 20) {
  return parseCKPriceChangeListings(
    payload?.bottom,
    (priceChangeUsd) => priceChangeUsd < 0,
    limit,
  );
}

export function hasNonZeroUsdPriceChanges(increases, drops) {
  return [...increases, ...drops].some(
    (item) =>
      typeof item?.priceChangeUsd === "number" &&
      Number.isFinite(item.priceChangeUsd) &&
      item.priceChangeUsd !== 0,
  );
}
