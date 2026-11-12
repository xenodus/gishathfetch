import { normalizeExtraInfo } from "./cardIdentity";

export function savedCardMatchesFilter(card, query) {
  const normalizedQuery = String(query ?? "")
    .trim()
    .toLowerCase();
  if (!normalizedQuery) {
    return true;
  }

  const fields = [
    card?.name,
    card?.src,
    card?.quality,
    normalizeExtraInfo(card?.extraInfo),
  ];

  if (card?.isFoil) {
    fields.push("foil");
  }

  return fields.some((value) => {
    if (value == null || value === "") {
      return false;
    }
    return String(value).toLowerCase().includes(normalizedQuery);
  });
}
