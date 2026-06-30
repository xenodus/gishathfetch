const normalizeText = (value) =>
  String(value ?? "")
    .trim()
    .normalize("NFKC")
    .replace(/[\u2018\u2019\u0060]/g, "'");

export const normalizeExtraInfo = (extraInfo) => {
  if (extraInfo == null) {
    return "";
  }

  const text = Array.isArray(extraInfo)
    ? extraInfo.join(" ")
    : String(extraInfo);
  const trimmed = normalizeText(text);
  if (!trimmed) {
    return "";
  }

  const bracketParts = [...trimmed.matchAll(/\[([^\]]+)\]/g)].map(
    (match) => `[${match[1].trim()}]`,
  );
  if (bracketParts.length > 0) {
    return [...new Set(bracketParts)].sort().join(" ");
  }

  return trimmed;
};

export const cardIdentityKey = (card) =>
  JSON.stringify({
    src: normalizeText(card.src),
    name: normalizeText(card.name),
    extraInfo: normalizeExtraInfo(card.extraInfo),
    quality: normalizeText(card.quality),
    isFoil: !!card.isFoil,
  });

export const cardsExactMatch = (a, b) =>
  cardIdentityKey(a) === cardIdentityKey(b);

export const dedupeCartItems = (items) => {
  const seen = new Set();
  const deduped = [];

  for (const item of items) {
    const key = cardIdentityKey(item);
    if (seen.has(key)) {
      continue;
    }
    seen.add(key);
    deduped.push(item);
  }

  return deduped;
};
