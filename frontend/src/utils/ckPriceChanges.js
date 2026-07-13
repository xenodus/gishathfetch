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
