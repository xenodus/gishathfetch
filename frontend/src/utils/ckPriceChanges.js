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

function parseCKPriceChangeListings(listings, matchesDirection, limit = 20) {
  if (!Array.isArray(listings)) {
    return [];
  }

  return listings
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
      };
    })
    .filter(
      (item) =>
        item.cardName.length > 0 &&
        item.priceChangeUsd !== null &&
        matchesDirection(item.priceChangeUsd),
    )
    .slice(0, limit);
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
