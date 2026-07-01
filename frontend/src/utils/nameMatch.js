export const NAME_MATCH_TIER = {
  EXACT: 0,
  PREFIX: 1,
  PARTIAL: 2,
  NONE: 3,
};

export function getNameMatchTier(cardName, searchQuery) {
  const lowerName = String(cardName ?? "").toLowerCase();
  const lowerQuery = String(searchQuery ?? "")
    .trim()
    .toLowerCase();

  if (!lowerQuery || !lowerName.includes(lowerQuery)) {
    return NAME_MATCH_TIER.NONE;
  }
  if (lowerName === lowerQuery) {
    return NAME_MATCH_TIER.EXACT;
  }
  if (lowerName.startsWith(lowerQuery)) {
    return NAME_MATCH_TIER.PREFIX;
  }
  return NAME_MATCH_TIER.PARTIAL;
}
