export function formatPriceChangePercent(percent, { absolute = false } = {}) {
  if (typeof percent !== "number" || !Number.isFinite(percent)) {
    return null;
  }

  const rounded = Math.round(percent * 10) / 10;
  const value = absolute ? Math.abs(rounded) : rounded;
  return Number.isInteger(value) ? `${value}%` : `${value.toFixed(1)}%`;
}

function parseCKPriceChangeListings(listings, limit = 20) {
  if (!Array.isArray(listings)) {
    return [];
  }

  return listings
    .map((item) => ({
      cardName: typeof item?.cardName === "string" ? item.cardName.trim() : "",
      priceChangePercent:
        typeof item?.priceChangePercent === "number"
          ? item.priceChangePercent
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
