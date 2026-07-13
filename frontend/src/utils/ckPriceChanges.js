export function formatPriceChangePercent(percent) {
  if (typeof percent !== "number" || !Number.isFinite(percent)) {
    return null;
  }

  const rounded = Math.round(percent * 10) / 10;
  return Number.isInteger(rounded) ? `${rounded}%` : `${rounded.toFixed(1)}%`;
}

export function parseCKPriceIncreases(payload, limit = 20) {
  if (!Array.isArray(payload?.top)) {
    return [];
  }

  return payload.top
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
