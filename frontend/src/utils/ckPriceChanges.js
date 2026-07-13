export function formatPriceChangeUsd(usd) {
  if (typeof usd !== "number" || !Number.isFinite(usd)) {
    return null;
  }

  const rounded = Math.round(usd * 100) / 100;
  const amount = Math.abs(rounded).toFixed(2);

  if (rounded > 0) {
    return `+$${amount}`;
  }
  if (rounded < 0) {
    return `-$${amount}`;
  }
  return `$${amount}`;
}

function parseCKPriceChangeListings(listings, limit = 20) {
  if (!Array.isArray(listings)) {
    return [];
  }

  return listings
    .map((item) => ({
      cardName: typeof item?.cardName === "string" ? item.cardName.trim() : "",
      priceChangeUsd:
        typeof item?.priceChangeUsd === "number" &&
        Number.isFinite(item.priceChangeUsd)
          ? item.priceChangeUsd
          : null,
    }))
    .filter((item) => item.cardName.length > 0)
    .slice(0, limit);
}

export function parseCKPriceIncreases(payload, limit = 20) {
  return parseCKPriceChangeListings(payload?.top, limit);
}

export function parseCKPriceDrops(payload, limit = 20) {
  return parseCKPriceChangeListings(payload?.bottom, limit);
}
